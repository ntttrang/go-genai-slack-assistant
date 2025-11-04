# CI/CD Setup Guide: Separate CI and CD Pipelines with Multi-Level Approval

This guide explains how to configure Jenkins and GitHub to implement separated CI/CD pipelines with a robust approval workflow where:
- **CI stages** run automatically on every Pull Request
- **CD stages** run only after PR is merged to `develop` or `main` branches
- GitHub blocks PR merges until all CI checks pass
- **Multi-level approval workflow** enforces code review by Reviewer → Approver → Owner before merge
- Code owners must approve changes to their domains
- All review conversations must be resolved before merging

## Table of Contents
1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [GitHub Webhook Configuration](#github-webhook-configuration)
4. [Jenkins Multibranch Pipeline Setup](#jenkins-multibranch-pipeline-setup)
5. [GitHub Branch Protection Rules](#github-branch-protection-rules)
6. [Testing the Setup](#testing-the-setup)
7. [Troubleshooting](#troubleshooting)

## Overview

### Pipeline Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Developer Workflow                        │
└─────────────────────────────────────────────────────────────┘

1. Developer creates feature branch
2. Developer pushes commits and creates Pull Request
3. GitHub webhook triggers Jenkins
4. Jenkins runs CI stages (1-9):
   ├── Checkout
   ├── Environment Setup
   ├── Dependencies
   ├── Lint
   ├── Security Scans (Gosec, Govulncheck, Trivy)
   ├── Unit Tests
   ├── Quality Analysis (SonarCloud)
   ├── Integration Tests
   └── Build
5. Jenkins reports status to GitHub
6. If CI fails → Merge blocked ❌ (stop here)
7. If CI passes → Review & Approval workflow begins:
   ├── 👤 Reviewer reviews code and leaves comments
   ├── 👤 Approver reviews and addresses reviewer's comments
   ├── 👤 Owner performs final verification
   └── ✅ Required approvals met → Merge button enabled
8. After merge to develop/main → CD stages run (10-14):
   ├── Build Docker Image
   ├── Push to Docker Hub
   ├── Deploy to Staging/Production
   └── Health Check
```

## Prerequisites

Before starting, ensure you have:

- [ ] Jenkins server installed and running (v2.300+)
- [ ] GitHub repository with admin access
- [ ] GitHub Personal Access Token with `repo` and `admin:repo_hook` scopes
- [ ] Jenkins plugins installed:
  - GitHub Branch Source Plugin
  - GitHub Plugin
  - Pipeline Plugin
  - Pipeline: Multibranch Plugin
  - Credentials Plugin
  - Slack Notification Plugin (optional)

### Generate GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Set token name: `Jenkins CI/CD`
4. Select scopes:
   - ✅ `repo` (Full control of private repositories)
   - ✅ `admin:repo_hook` (Full control of repository hooks)
5. Click "Generate token"
6. **Copy the token immediately** (you won't see it again)

## GitHub Webhook Configuration

### Step 1: Add Webhook to GitHub Repository

1. Navigate to your GitHub repository
2. Go to **Settings** → **Webhooks** → **Add webhook**

3. Configure webhook settings:
   ```
   Payload URL: https://your-jenkins-server.com/github-webhook/
   Content type: application/json
   Secret: (optional, but recommended for security)
   ```

4. Select events to trigger webhook:
   - ✅ **Pushes** - Triggers builds when code is pushed
   - ✅ **Pull requests** - Triggers CI when PRs are created/updated
   - ✅ **Pull request reviews** - (optional) Triggers on PR reviews

5. Ensure "Active" is checked

6. Click **Add webhook**

### Step 2: Verify Webhook

1. After saving, GitHub will send a test ping
2. Check the "Recent Deliveries" tab
3. Verify response shows `200 OK`

**Troubleshooting Webhook Issues:**
- If you get connection errors, ensure Jenkins is accessible from internet
- For local Jenkins, use ngrok or similar tunneling service
- Check Jenkins logs: `Manage Jenkins → System Log → github-webhook`

## Jenkins Multibranch Pipeline Setup

### Step 1: Add GitHub Credentials to Jenkins

1. Go to Jenkins → **Manage Jenkins** → **Manage Credentials**
2. Select domain `(global)`
3. Click **Add Credentials**
4. Configure:
   ```
   Kind: Username with password
   Scope: Global
   Username: your-github-username
   Password: <paste-your-github-token>
   ID: github-credentials
   Description: GitHub Token for CI/CD
   ```
5. Click **OK**

### Step 2: Create Multibranch Pipeline Job

1. From Jenkins dashboard, click **New Item**
2. Enter name: `go-genai-slack-assistant`
3. Select **Multibranch Pipeline**
4. Click **OK**

### Step 3: Configure Branch Sources

1. Under **Branch Sources**, click **Add source** → **GitHub**

2. Configure GitHub source:
   ```
   Credentials: github-credentials (select from dropdown)
   Repository HTTPS URL: https://github.com/ntttrang/go-genai-slack-assistant.git
   ```

3. Under **Behaviors**, click **Add** and configure:

   **a) Discover branches**
   - Strategy: `All branches`

   **b) Discover pull requests from origin**
   - Strategy: `Merging the pull request with the current target branch revision`
   - This creates a merged version of the PR for testing

   **c) Filter by name (with regular expression)** (optional)
   - Regular expression: `(main|develop|feature/.*|bugfix/.*)`
   - This limits which branches trigger builds

### Step 4: Configure Build Configuration

1. Under **Build Configuration**:
   ```
   Mode: by Jenkinsfile
   Script Path: Jenkinsfile
   ```

### Step 5: Configure Scan Triggers

1. Under **Scan Multibranch Pipeline Triggers**:
   - ✅ Check "Periodically if not otherwise run"
   - Interval: `1 minute` (fallback if webhook fails)

2. Under **Orphaned Item Strategy**:
   ```
   Days to keep old items: 7
   Max # of old items to keep: 20
   ```

### Step 6: Save Configuration

1. Click **Save**
2. Jenkins will immediately scan repository
3. It will:
   - Discover all branches matching filters
   - Discover all open pull requests
   - Create sub-jobs for each branch/PR
   - Trigger initial builds

## GitHub Branch Protection Rules

Branch protection rules enforce quality gates by blocking merges until CI passes and required approvals are obtained.

### Step 1: Set Up CODEOWNERS File (Required for Multi-Level Approvals)

Before configuring branch protection, create a CODEOWNERS file to define who must approve changes:

1. Create `.github/CODEOWNERS` file in your repository root:
   ```bash
   mkdir -p .github
   touch .github/CODEOWNERS
   ```

2. Define code ownership rules (example):
   ```
   # Default owners for everything in the repo
   # At least one owner must approve
   * @owner-username
   
   # Backend code requires approval from backend team
   /internal/** @backend-reviewer @backend-approver
   /pkg/** @backend-reviewer @backend-approver
   
   # Infrastructure changes require DevOps approval
   /infrastructure/** @devops-team @senior-devops
   /Jenkinsfile @devops-team
   /docker-compose.yml @devops-team
   
   # Documentation can be approved by tech writers
   /docs/** @tech-writer @owner-username
   ```

3. Commit and push the CODEOWNERS file:
   ```bash
   git add .github/CODEOWNERS
   git commit -m "feat: add CODEOWNERS for multi-level approval workflow"
   git push origin main
   ```

### Step 2: Configure Protection for `develop` Branch

1. Go to GitHub repository → **Settings** → **Branches**
2. Under "Branch protection rules", click **Add branch protection rule**

3. Configure rule for `develop`:

   **Branch name pattern:**
   ```
   develop
   ```

   **Protection settings:**

   - ✅ **Require a pull request before merging**
     - ✅ Require approvals: `2` or `3` (for multi-level approval: reviewer + approver + owner)
     - ✅ Dismiss stale pull request approvals when new commits are pushed
     - ✅ **Require review from Code Owners** ⭐ (critical for multi-level workflow)
     - ⬜ Require approval of the most recent reviewable push

   - ✅ **Require status checks to pass before merging**
     - ✅ Require branches to be up to date before merging
     - **Status checks that are required:**
       - Search and select: `continuous-integration/jenkins/pr-merge`
       - Alternative: `continuous-integration/jenkins/branch`
       - Or: `go-genai-slack-assistant` (your Jenkins job name)
       
       > **Note:** Status checks only appear after first build. Create a test PR first, then come back to add the check.

   - ✅ **Require conversation resolution before merging** ⭐ (ensures all review comments are addressed)

   - ✅ **Require linear history** (optional, prevents merge commits)

   - ✅ **Do not allow bypassing the above settings**
     - ⬜ Uncheck "Allow specified actors to bypass required pull requests"
     - This ensures even admins follow the approval workflow

4. Click **Create** or **Save changes**

### Multi-Level Approval Workflow Explanation

With the above configuration, the approval flow works as follows:

```
┌────────────────────────────────────────────────────────┐
│         Multi-Level Approval Process                    │
└────────────────────────────────────────────────────────┘

Step 1: CI Must Pass First
├─ ❌ If CI fails → PR blocked, no reviews accepted yet
└─ ✅ If CI passes → Approval workflow can begin

Step 2: Reviewer Reviews Code
├─ Reviewer examines code changes
├─ Leaves comments and suggestions
└─ Can "Request changes" to block merge

Step 3: Approver Reviews
├─ Approver checks reviewer's comments
├─ Verifies reviewer's concerns are addressed
└─ Provides approval (1st approval)

Step 4: Owner Final Check
├─ Code owner performs final verification
├─ Ensures all standards met
└─ Provides approval (2nd/3rd approval)

Step 5: Merge Enabled
└─ ✅ All required approvals + CI passed + conversations resolved
   → Merge button enabled
```

**Key Points:**
- **CI runs first** - No point in reviewing code that doesn't pass tests
- **Code Owners must approve** - Ensures domain experts review relevant changes
- **Conversation resolution required** - All review comments must be addressed
- **Stale approvals dismissed** - New commits require re-approval, preventing sneaky changes
- **No bypass allowed** - Even repository admins must follow the process

### Step 3: Configure Protection for `main` Branch

Repeat the same steps for the `main` branch with **stricter rules for production**:

**Branch name pattern:**
```
main
```

**Protection settings (stricter for production):**

- ✅ **Require a pull request before merging**
  - ✅ Require approvals: `3` or more (higher for production: reviewer + approver + owner)
  - ✅ Dismiss stale pull request approvals when new commits are pushed
  - ✅ **Require review from Code Owners**
  - ✅ Require approval of the most recent reviewable push

- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - Required status checks:
    - `continuous-integration/jenkins/branch` (full CI/CD pipeline)
    - Consider adding: deployment to staging successful
    - Consider adding: security scan passed

- ✅ **Require conversation resolution before merging**

- ✅ **Require linear history** (recommended for production)

- ✅ **Require deployments to succeed before merging** (optional)
  - Select: staging environment

- ✅ **Restrict who can push to matching branches** (optional, highly recommended)
  - Select specific users/teams: DevOps team, Tech leads, Release managers
  - This prevents direct pushes, enforcing PR workflow

- ✅ **Do not allow bypassing the above settings**
  - ⬜ Uncheck "Allow specified actors to bypass required pull requests"

**Recommended CODEOWNERS for `main` branch:**
```
# For main branch, require senior approval
* @tech-lead @devops-lead @owner-username
```

### Summary: What You've Configured

After completing Steps 1-3, your repository will have:

**✅ Automated CI Pipeline**
- Every PR triggers Jenkins CI automatically
- CI must pass before any reviews or approvals

**✅ Multi-Level Approval Enforcement**
- 2+ approvals required for develop branch
- 3+ approvals required for main branch
- Code owners MUST approve changes to their code
- Stale approvals dismissed on new commits

**✅ Quality Gates**
- All conversations must be resolved
- Branches must be up to date before merging
- Status checks must pass
- Linear history maintained (optional)

**✅ Merge Protection**
- Direct pushes blocked (PR workflow enforced)
- Admins cannot bypass rules
- Deployment gates for production (optional)

**Result:** A robust, secure approval workflow that ensures code quality, security, and team collaboration before any code reaches your branches.

### Understanding Status Checks

Jenkins reports build status to GitHub using these formats:

**For Pull Requests:**
- Status check name: `continuous-integration/jenkins/pr-merge`
- Shows CI stages (1-9) results
- Must pass before merge is allowed

**For Branch Builds:**
- Status check name: `continuous-integration/jenkins/branch`
- Shows full CI/CD pipeline results

**Status Types:**
- ✅ Success - All checks passed, merge allowed
- ❌ Failure - Checks failed, merge blocked
- 🟡 Pending - Build in progress
- ⚪ Expected - Waiting for status

## Testing the Setup

Follow these steps to verify the CI/CD separation works correctly:

### Test 1: Verify PR-based CI

1. Create a feature branch:
   ```bash
   git checkout -b feature/test-ci-pipeline
   ```

2. Make a small change (e.g., add comment to README):
   ```bash
   echo "# Testing CI pipeline" >> README.md
   git add README.md
   git commit -m "test: verify CI pipeline on PR"
   ```

3. Push branch and create Pull Request:
   ```bash
   git push origin feature/test-ci-pipeline
   ```
   Then create PR on GitHub targeting `develop`

4. **Expected behavior:**
   - Jenkins automatically detects new PR
   - Triggers build within 1 minute
   - Runs CI stages 1-9:
     - ✅ Checkout
     - ✅ Environment Setup
     - ✅ Dependencies
     - ✅ Lint
     - ✅ Security Scans
     - ✅ Unit Tests
     - ✅ Quality Analysis
     - ✅ Integration Tests
     - ✅ Build
   - **CD stages 10-14 are SKIPPED** with message:
     ```
     Stage 'Build Docker Image' skipped due to when conditional
     Stage 'Push to Docker Hub' skipped due to when conditional
     Stage 'Deploy to Staging' skipped due to when conditional
     ...
     ```
   - GitHub shows status check on PR
   - Merge button is:
     - ✅ **Enabled** if CI passes
     - ❌ **Blocked** if CI fails

5. Check Jenkins console output for:
   ```
   Building Pull Request #1
   PR Title: Verify CI pipeline on PR
   PR Author: your-username
   Target Branch: develop
   ```

6. Verify Slack notification (if configured):
   ```
   ✅ CI SUCCESS - Pull Request
   PR #1: Verify CI pipeline on PR
   Author: your-username
   Target: develop
   ```

### Test 2: Verify Post-Merge CD

1. After CI passes, merge the Pull Request on GitHub

2. **Expected behavior:**
   - Jenkins detects merge to `develop`
   - Triggers new build for `develop` branch
   - Runs **ALL stages 1-14**:
     - CI stages 1-9 (same as PR)
     - **CD stages 10-14 now execute:**
       - ✅ Build Docker Image
       - ✅ Push to Docker Hub
       - ✅ Deploy to Staging
       - ✅ Health Check

3. Check Jenkins console output for:
   ```
   Building Branch: develop
   CD STAGE 10: BUILD DOCKER IMAGE
   CD STAGE 11: PUSH TO DOCKER HUB
   CD STAGE 12: DEPLOY TO STAGING (RENDER)
   CD STAGE 14: HEALTH CHECK
   ```

4. Verify Slack notification:
   ```
   ✅ CI/CD SUCCESS
   Branch: develop
   Environment: staging
   ```

5. Verify deployment:
   - Check Docker Hub for new image tag
   - Visit staging URL to confirm deployment
   - Check health endpoint returns 200

### Test 3: Verify Merge Blocking

1. Create a failing test:
   ```bash
   git checkout -b feature/test-merge-blocking
   # Make a change that breaks tests
   git push origin feature/test-merge-blocking
   ```

2. Create Pull Request to `develop`

3. **Expected behavior:**
   - Jenkins runs CI stages
   - CI fails (e.g., test failure)
   - GitHub status shows ❌ Failure
   - **Merge button is disabled** with message:
     ```
     Merging is blocked
     Required status check "continuous-integration/jenkins/pr-merge" has failed
     ```

4. Fix the issue, push new commit

5. Jenkins automatically re-runs CI

6. After CI passes, merge button becomes enabled

### Test 4: Verify Multi-Level Approval Workflow

This test verifies the complete approval chain: CI → Reviewer → Approver → Owner → Merge.

1. Create a feature branch with substantial changes:
   ```bash
   git checkout -b feature/test-approval-workflow
   # Make meaningful code changes
   echo "// New feature implementation" >> internal/app/handler.go
   git add internal/app/handler.go
   git commit -m "feat: add new feature requiring multi-level approval"
   git push origin feature/test-approval-workflow
   ```

2. Create Pull Request to `develop`

3. **Expected behavior - Stage 1: CI Check**
   - Jenkins automatically triggers CI stages 1-9
   - GitHub shows "Checks running" status
   - Merge button shows:
     ```
     Merging is blocked
     - Required status check is in progress
     - At least 2 approving reviews are required
     ```

4. **Expected behavior - Stage 2: CI Pass, Awaiting Reviews**
   - CI completes successfully ✅
   - Merge button shows:
     ```
     Merging is blocked
     - Review required from code owners
     - At least 2 approving reviews are required by reviewers with write access
     ```

5. **Test Scenario A: Reviewer reviews first**
   - Reviewer (not a code owner) adds review comments
   - Reviewer clicks "Comment" (NOT "Approve" yet)
   - Merge button remains blocked:
     ```
     Merging is blocked
     - At least 2 approving reviews are required
     ```

6. **Test Scenario B: Try to merge without enough approvals**
   - One person approves
   - Merge button still shows:
     ```
     Merging is blocked
     - At least 1 more approving review is required
     ```

7. **Test Scenario C: Code Owner approves**
   - Code owner (from CODEOWNERS file) provides approval
   - Check that "Required review from code owners" ✅ turns green
   - If only 1 approval exists, merge still blocked until 2nd approval

8. **Test Scenario D: All approvals met**
   - Second approver provides approval
   - All checks turn green:
     ```
     ✅ Required reviews approved
     ✅ Review required from code owners
     ✅ All checks have passed
     ```
   - **Merge button becomes enabled** 🎉

9. **Test Scenario E: New commit resets approvals**
   - Before merging, push a new commit
   - Watch approvals get dismissed automatically
   - Merge button blocks again:
     ```
     Merging is blocked
     - At least 2 approving reviews are required
     ```
   - Reviewers must re-approve

10. **Test Scenario F: Unresolved conversations block merge**
    - Have a reviewer leave a comment
    - Try to get all approvals
    - Even with approvals, merge shows:
      ```
      Merging is blocked
      - 1 unresolved conversation
      ```
    - Resolve the conversation
    - Merge button enables

11. **Verify complete flow:**
    ```
    ✅ CI passed
    ✅ 2+ approvals obtained
    ✅ Code owner approved
    ✅ All conversations resolved
    → Merge button enabled → Successfully merge PR
    ```

**Key Verification Points:**
- [ ] CI must pass before meaningful reviews
- [ ] Code owners MUST approve (can't be bypassed)
- [ ] Multiple approvals required (2 for develop, 3 for main)
- [ ] New commits dismiss previous approvals
- [ ] Unresolved conversations block merge
- [ ] Even admins cannot bypass (if configured correctly)

## Troubleshooting

### Issue: Jenkins doesn't detect Pull Requests

**Symptoms:**
- New PRs don't trigger builds
- Only branch builds work

**Solutions:**
1. Check "Discover pull requests from origin" behavior is enabled
2. Verify webhook events include "Pull requests"
3. Manually trigger scan: Jenkins job → "Scan Multibranch Pipeline Now"
4. Check Jenkins logs: `Manage Jenkins → System Log → Add new recorder → "GitHub"`

### Issue: GitHub doesn't show status checks

**Symptoms:**
- Jenkins builds successfully
- No status appears on PR

**Solutions:**
1. Verify GitHub credentials have `repo:status` permission
2. Check Jenkins GitHub plugin configuration:
   - `Manage Jenkins → Configure System → GitHub`
   - Verify GitHub Server is configured
   - Test connection
3. Ensure Jenkins is accessible from GitHub (not localhost)
4. Check "Advanced" settings in Multibranch Pipeline job

### Issue: Merge button not blocked despite failed CI

**Symptoms:**
- CI fails but merge is still allowed

**Solutions:**
1. Go to branch protection rules
2. Ensure "Require status checks to pass before merging" is checked
3. Add required status check (appears after first build)
4. Ensure "Do not allow bypassing" is enabled
5. Check if you're a repository admin (admins can bypass by default)

### Issue: CD stages run on Pull Requests

**Symptoms:**
- Docker images built on PR
- Deployment attempted from PR

**Solutions:**
1. Verify `when` conditions in Jenkinsfile:
   ```groovy
   when {
       allOf {
           not { changeRequest() }
           branch 'develop'
       }
   }
   ```
2. Check Jenkins console shows "Stage skipped due to when conditional"
3. Ensure using Multibranch Pipeline (not freestyle job)

### Issue: Webhook returns 403 Forbidden

**Symptoms:**
- GitHub shows 403 errors in webhook deliveries

**Solutions:**
1. Enable "GitHub hook trigger for GITScm polling" in Jenkins job
2. Configure GitHub server in Jenkins:
   - `Manage Jenkins → Configure System → GitHub`
   - Add GitHub Server
   - Test credentials
3. Ensure webhook URL is correct: `https://jenkins-url/github-webhook/`
4. Check firewall rules allow GitHub IPs

### Issue: "Expected" status never updates

**Symptoms:**
- PR shows status as "Expected — Waiting for status to be reported"

**Solutions:**
1. Jenkins may not have reported status yet
2. Check if build started in Jenkins
3. Manually trigger scan or build
4. Remove status check from branch protection, trigger build, then re-add

### Issue: Code owner approval not recognized

**Symptoms:**
- Code owner approves but "Review required from code owners" still shows as pending
- Merge button remains blocked

**Solutions:**
1. Verify CODEOWNERS file is in correct location: `.github/CODEOWNERS`
2. Check CODEOWNERS syntax - use GitHub usernames with `@` prefix
3. Ensure code owner has write access to the repository
4. CODEOWNERS changes only take effect on new PRs (not existing ones)
5. Test with a fresh PR after updating CODEOWNERS
6. Verify code owner is approving the correct PR target branch

### Issue: Merge button enabled with insufficient approvals

**Symptoms:**
- Merge allowed with only 1 approval when 2+ required
- Admins can bypass protection rules

**Solutions:**
1. Check "Do not allow bypassing" is enabled in branch protection
2. Verify you're not a repository admin (admins can bypass by default)
3. Ensure "Require approvals" number is correctly set
4. Check if "Allow specified actors to bypass" has exceptions listed
5. Branch protection rules may not have saved - re-apply and save again

### Issue: Approvals dismissed unexpectedly

**Symptoms:**
- Approvals disappear after minor changes
- Have to re-request approvals frequently

**Expected Behavior:**
- This is correct! "Dismiss stale approvals when new commits are pushed" is enabled
- Any new commit resets approvals to prevent unauthorized changes
- This is a security feature, not a bug

**Solutions:**
1. Only push commits when ready for re-review
2. Use draft PRs for work-in-progress changes
3. Squash/amend commits before requesting final approval
4. Coordinate with reviewers to approve after final changes

## Best Practices

### 1. Secure Your Setup
- Use HTTPS for Jenkins (required for GitHub webhooks)
- Enable CSRF protection in Jenkins
- Use webhook secrets for authentication
- Restrict Jenkins credentials access
- Regularly rotate GitHub tokens

### 2. Optimize Build Performance
- Use Jenkins agents for parallel builds
- Cache dependencies (Go modules, Docker layers)
- Skip unnecessary stages on documentation-only changes
- Use fast-failing stages early in pipeline

### 3. Clear Communication
- Use descriptive commit messages (conventional commits)
- Add comments explaining complex when conditions
- Document custom pipeline stages
- Maintain this setup guide

### 4. Monitor and Maintain
- Set up Jenkins monitoring and alerts
- Review failed builds regularly
- Keep Jenkins and plugins updated
- Audit branch protection rules quarterly
- Monitor webhook delivery success rate

### 5. Developer Experience
- Provide clear CI failure messages
- Fast feedback loops (< 10 minutes for CI)
- Slack/email notifications for build status
- Dashboard showing build metrics
- Documentation for resolving common failures

### 6. Multi-Level Approval Best Practices
- **Define clear roles:** Document who are reviewers, approvers, and owners
- **CODEOWNERS hygiene:** Keep CODEOWNERS file up-to-date with team changes
- **Appropriate approval counts:** 
  - Develop branch: 2 approvals (reviewer + approver)
  - Main branch: 3 approvals (reviewer + approver + owner/tech-lead)
- **Review efficiently:** 
  - Reviewers focus on code quality, logic, tests
  - Approvers verify reviewer feedback is addressed
  - Owners ensure architectural alignment
- **Use draft PRs:** Mark work-in-progress PRs as drafts to avoid premature reviews
- **Batch changes:** Squash commits before final approval to avoid re-approval cycles
- **Communicate:** Use PR comments to explain complex changes upfront
- **Set SLAs:** Define expected review turnaround times (e.g., 24 hours for reviews)
- **Avoid bottlenecks:** Have backup code owners to prevent single points of failure
- **Emergency process:** Document expedited approval process for critical hotfixes

## Related Documentation

- [Jenkinsfile](../Jenkinsfile) - Pipeline definition
- [SLACK_SETUP.md](./SLACK_SETUP.md) - Slack integration
- [README.md](../README.md) - Project overview
- [Jenkins Official Docs](https://www.jenkins.io/doc/book/pipeline/multibranch/)
- [GitHub Branch Protection](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches)

## Support

If you encounter issues not covered in this guide:

1. Check Jenkins console logs for detailed error messages
2. Review GitHub webhook delivery logs
3. Consult Jenkins system logs: `Manage Jenkins → System Log`
4. Check GitHub Actions/Jenkins plugin compatibility
5. Reach out to DevOps team for assistance

---

**Last Updated:** 2025-11-04  
**Version:** 2.0 (Added Multi-Level Approval Workflow)  
**Maintainer:** DevOps Team

