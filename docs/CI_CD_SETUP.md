# CI/CD Setup Guide: Separate CI and CD Pipelines

This guide explains how to configure Jenkins and GitHub to implement separated CI/CD pipelines where:
- **CI stages** run automatically on every Pull Request
- **CD stages** run only after PR is merged to `develop` or `main` branches
- GitHub blocks PR merges until all CI checks pass

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
6. If CI passes → Merge button enabled
7. If CI fails → Merge blocked ❌
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

Branch protection rules enforce quality gates by blocking merges until CI passes.

### Step 1: Configure Protection for `develop` Branch

1. Go to GitHub repository → **Settings** → **Branches**
2. Under "Branch protection rules", click **Add branch protection rule**

3. Configure rule for `develop`:

   **Branch name pattern:**
   ```
   develop
   ```

   **Protection settings:**

   - ✅ **Require a pull request before merging**
     - ✅ Require approvals: `1` (or more)
     - ✅ Dismiss stale pull request approvals when new commits are pushed
     - ✅ Require review from Code Owners (optional)

   - ✅ **Require status checks to pass before merging**
     - ✅ Require branches to be up to date before merging
     - **Status checks that are required:**
       - Search and select: `continuous-integration/jenkins/branch`
       - Or: `go-genai-slack-assistant` (your Jenkins job name)
       
       > **Note:** Status checks only appear after first build. Create a test PR first, then come back to add the check.

   - ✅ **Require conversation resolution before merging** (optional)

   - ✅ **Require linear history** (optional, prevents merge commits)

   - ✅ **Do not allow bypassing the above settings**
     - Uncheck "Allow specified actors to bypass required pull requests"

4. Click **Create** or **Save changes**

### Step 2: Configure Protection for `main` Branch

Repeat the same steps for the `main` branch with potentially stricter rules:

```
Branch name pattern: main

Additional settings for production:
- Require approvals: 2 (higher for production)
- Require deployment to staging before production (optional)
- Restrict who can push to matching branches (optional)
  - Select: DevOps team, Tech leads
```

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
**Version:** 1.0  
**Maintainer:** DevOps Team

