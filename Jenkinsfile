pipeline {
    agent any

    triggers {
        githubPush()
    }

    environment {
        GO_VERSION = '1.24.1'
        GOPATH = "${env.WORKSPACE}/.go"
        GOMODCACHE = "${env.WORKSPACE}/.gomodcache"
        GOCACHE = "${env.WORKSPACE}/.gocache"
        GOTOOLCHAIN = 'auto'
        CGO_ENABLED = '0'
        GOOS = 'linux'
        GOARCH = 'amd64'
        
        // Docker configuration
        DOCKER_REGISTRY = 'https://index.docker.io/v1/'
        DOCKER_IMAGE_NAME = 'docker.io/minhtrang2106/slack-bot'
        
        // Render configuration
        RENDER_API_KEY_CREDENTIALS_ID = 'render-api-key'  // Jenkins credential ID for Render API key
        RENDER_STAGING_SERVICE_ID = 'srv-d427l5k9c44c7385bou0'
        RENDER_PRODUCTION_SERVICE_ID = 'srv-yyyyy'
        RENDER_STAGING_URL = 'https://slack-bot-63-a2ae2ba.onrender.com'
        RENDER_PRODUCTION_URL = 'https://slack-bot-production.onrender.com'

        // Slack Notification
        SLACK_CHANNEL = '#jenkins-cicd'
        SLACK_BOT_TOKEN = 'SLACK_BOT_TOKEN'

        SONAR_TOKEN = 'sonarcloud-token'

    }

    stages {
        // ===============================
        // CI STAGES
        // ===============================
        stage('1. Checkout') {
            steps {
                deleteDir()
                echo '============================================'
                echo 'CI STAGE 1: CHECKOUT SOURCE CODE'
                echo '============================================'
                echo 'Checking out repository...'
                git(
                    url: 'https://github.com/ntttrang/go-genai-slack-assistant.git',
                    branch: 'develop',
                    credentialsId: 'github-credentials'
                )
                script {
                    def gitBranch = sh(returnStdout: true, script: 'git rev-parse --abbrev-ref HEAD').trim()
                    def gitCommit = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
                    echo "Branch: ${gitBranch}"
                    echo "Commit: ${gitCommit}"
                    
                    // Set BRANCH_NAME for when conditions to work
                    env.BRANCH_NAME = gitBranch
                    echo "BRANCH_NAME set to: ${env.BRANCH_NAME}"
                }
            }
        }
        stage('2. Environment Setup') {
            steps {
                echo '============================================'
                echo 'CI STAGE 2: ENVIRONMENT SETUP'
                echo '============================================'
                echo 'Setting up Go environment...'
                sh '''
                    go version
                    go env
                    
                    # Setup fresh Go cache directories
                    echo "Setting up Go cache directories..."
                    mkdir -p "${WORKSPACE}/.gomodcache" "${WORKSPACE}/.gocache"
                    export GOPATH="${WORKSPACE}/.go"
                    mkdir -p "${GOPATH}/bin" "${GOPATH}/pkg"
                    chmod -R 755 "${GOPATH}" "${WORKSPACE}/.gomodcache" "${WORKSPACE}/.gocache" || true
                '''
            }
        }

        stage('3. Dependencies') {
            steps {
                echo '============================================'
                echo 'CI STAGE 3: DOWNLOAD DEPENDENCIES'
                echo '============================================'
                echo 'Downloading and verifying dependencies...'
                sh '''
                    go mod download
                    go mod verify
                    go mod tidy
                    echo 'Dependencies check completed successfully'
                '''
            }
        }

        stage('Parallel Quality Checks') {
            parallel {
                stage('4. Lint') {
                    steps {
                        echo '============================================'
                        echo 'CI STAGE 4: CODE QUALITY CHECK'
                        echo '============================================'
                        echo 'Running code quality checks with golangci-lint...'
                        sh '''
                            which golangci-lint > /dev/null || (echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
                            golangci-lint run  --out-format=json ./... > golangci-lint-report.json || true
                            echo "Golangci-lint scan completed"
                        '''
                    }
                }

                stage('5.1 Gosec Scan') {
                    steps {
                        echo '============================================'
                        echo 'SECURITY SCAN: Gosec'
                        echo '============================================'
                        echo 'Running security scan with gosec...'
                        sh '''
                            export PATH="${GOPATH}/bin:${PATH}"
                            which gosec > /dev/null || (echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest)
                            gosec -fmt=json -out=gosec-report.json -exclude-dir=.gomodcache -exclude-dir=.go -exclude-dir=.gocache ./... || true
                            echo "Gosec scan completed"
                        '''
                    }
                }

                stage('5.2 Govulncheck Scan') {
                    steps {
                        echo '============================================'
                        echo 'SECURITY SCAN: Govulncheck'
                        echo '============================================'
                        echo 'Running vulnerability scan with govulncheck...'
                        sh '''
                            export PATH="${GOPATH}/bin:${PATH}"
                            which govulncheck > /dev/null || (echo "Installing govulncheck..."; go install golang.org/x/vuln/cmd/govulncheck@latest)
                            govulncheck -json ./... > govulncheck-report.json 2>&1 || true
                            echo "Govulncheck scan completed"
                        '''
                    }
                }

                stage('5.3 Trivy Scan') {
                    steps {
                        echo '============================================'
                        echo 'SECURITY SCAN: Trivy'
                        echo '============================================'
                        echo 'Running vulnerability scan with trivy...'
                        sh '''
                            export PATH="${WORKSPACE}/bin:${PATH}"
                            which trivy > /dev/null || (echo "Installing trivy..."; curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b ${WORKSPACE}/bin)
                            trivy fs --severity HIGH,CRITICAL --format json --output trivy-report.json --skip-dirs .gomodcache . || true
                            echo "Trivy scan completed"
                        '''
                    }
                }

                stage('6. Unit Test') {
                    steps {
                        echo '============================================'
                        echo 'CI STAGE 6: UNIT TESTS'
                        echo '============================================'
                        echo 'Running unit tests with coverage...'
                        sh '''
                            go test -v -coverprofile=coverage.out $(go list ./... | grep -v '/tests')
                            go tool cover -html=coverage.out -o coverage.html
                            echo "Unit tests completed"
                        '''
                    }
                }
            }
        }
        stage('7. Quality Analysis') {
            steps {
                echo '============================================'
                echo 'CI STAGE 7: QUALITY ANALYSIS: Sonar Scanner, SonarQube Cloud, Quality Gate'
                echo '============================================'
                script {
                      def scannerHome = tool 'SonarScanner'
                      withCredentials([string(credentialsId: env.SONAR_TOKEN, variable: 'SONAR_TOKEN')]) {
                            // This command executes the SonarScanner
                            sh "${scannerHome}/bin/sonar-scanner"
                        }

                        // Note: quality gate is available in SonarQube Cloud
                        // So we don't need to wait for quality gate
                        // We can just check the quality gate status in SonarQube Cloud
                        // And if it's not OK, we can fail the pipeline
                        // And if it's OK, we can continue the pipeline
                        // And if it's not OK, we can fail the pipeline
                        // !IMPORTANT: SonarCloud free tier can't use Webhook to notify Jenkins when analysis is complete
                }
            }
        }

        stage('8. Integration Test') {
            steps {
                echo '============================================'
                echo 'CI STAGE 7: INTEGRATION TESTS'
                echo '============================================'
                echo 'Running integration tests...'
                sh '''
                    go test -v -timeout=10m ./tests/...
                    echo "Integration tests completed"
                '''
            }
        }

        stage('9. Build') {
            steps {
                echo '============================================'
                echo 'CI STAGE 8: BUILD APPLICATION'
                echo '============================================'
                echo 'Building application...'
                sh '''
                    mkdir -p bin
                    go build -o bin/slack-bot cmd/api/main.go
                    echo "Build completed successfully"
                    ls -lah bin/slack-bot
                '''
            }
        }


        // ===============================
        // CD STAGES
        // ===============================
        stage('10. Build Docker Image') {
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 10: BUILD DOCKER IMAGE: docker build, Trivy Image Scan'
                    echo '============================================'
                    
                    def gitCommit = sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
                    def imageTag = "${env.BUILD_NUMBER}-${gitCommit}"
                    def fullImageName = "${DOCKER_IMAGE_NAME}:${imageTag}"
                    
                    echo "Building Docker image: ${fullImageName}"
                    sh """
                        docker build -t ${fullImageName} \
                            --platform linux/amd64 \
                            --build-arg BUILD_DATE=\$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
                            --build-arg VCS_REF=${gitCommit} \
                            -f Dockerfile .
                        docker tag ${fullImageName} ${DOCKER_IMAGE_NAME}:latest
                        echo "Docker image built successfully: ${fullImageName}"
                    """
                    
                    echo "Scanning Docker image for vulnerabilities with Trivy..."
                    sh """
                        export PATH="${WORKSPACE}/bin:\${PATH}"
                        trivy image --severity HIGH,CRITICAL \
                            --format json \
                            --output trivy-image-report.json \
                            ${fullImageName} || true
                        echo "Image scan completed"
                    """
                    
                    env.DOCKER_IMAGE = fullImageName
                    env.DOCKER_IMAGE_LATEST = "${DOCKER_IMAGE_NAME}:latest"
                    
                    echo "Docker image ready: ${env.DOCKER_IMAGE}"
                }
            }
        }

        stage('11. Push to Docker Hub') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 11: PUSH TO DOCKER HUB'
                    echo '============================================'
                    
                    withCredentials([usernamePassword(
                        credentialsId: 'dockerhub-credentials',
                        usernameVariable: 'DOCKER_USER',
                        passwordVariable: 'DOCKER_PASS'
                    )]) {
                        sh '''
                            echo "Logging into Docker Hub as ${DOCKER_USER}..."
                            echo "${DOCKER_PASS}" | docker login -u "${DOCKER_USER}" --password-stdin
                            
                            echo "Pushing image: ${DOCKER_IMAGE}"
                            docker push ${DOCKER_IMAGE}
                            
                            echo "Pushing latest tag: ${DOCKER_IMAGE_LATEST}"
                            docker push ${DOCKER_IMAGE_LATEST}
                            
                            echo "Successfully pushed images to Docker Hub"
                            docker logout
                        '''
                    }
                }
            }
        }

        stage('12. Deploy to Staging') {
            when {
                branch 'develop'
            }
            steps {
                script {
                    env.DEPLOY_ENVIRONMENT = 'staging'
                    echo '============================================'
                    echo 'CD STAGE 12: DEPLOY TO STAGING (RENDER)'
                    echo '============================================'
                    
                    withCredentials([
                        string(credentialsId: env.RENDER_API_KEY_CREDENTIALS_ID, variable: 'RENDER_API_KEY')
                    ]) {
                        sh """
                            echo "Triggering Render staging deployment..."
                            echo "Image: ${env.DOCKER_IMAGE}"
                            echo "Service ID: ${RENDER_STAGING_SERVICE_ID}"
                            
                            echo "Using Render API for deployment..."
                            curl -s -w "\\n%{http_code}" -X POST \
                            "https://api.render.com/v1/services/${RENDER_STAGING_SERVICE_ID}/deploys" \
                            -H "Authorization: Bearer \${RENDER_API_KEY}" \
                            -H "Content-Type: application/json" \
                            -d '{
                                "clearCache": "do_not_clear", "imageUrl": "'"${env.DOCKER_IMAGE}"'"}'
                            
                            echo "Staging deployment initiated successfully"
                            echo "Waiting 30 seconds for deployment to stabilize..."
                            sleep 30
                        """
                    }
                }
            }
        }

        stage('13. Deploy to Production') {
            when {
                branch 'main'
            }
            steps {
                script {
                    env.DEPLOY_ENVIRONMENT = 'production'
                    echo '============================================'
                    echo 'CD STAGE 13: DEPLOY TO PRODUCTION (RENDER)'
                    echo '============================================'
                    
                    timeout(time: 15, unit: 'MINUTES') {
                        input message: 'Deploy to Production?',
                              ok: 'Deploy',
                              submitter: 'admin,devops-team'
                    }
                    
                    withCredentials([
                        string(credentialsId: env.RENDER_API_KEY_CREDENTIALS_ID, variable: 'RENDER_API_KEY')
                    ]) {
                        sh """
                            echo "Triggering Render production deployment..."
                            echo "Image: ${env.DOCKER_IMAGE}"
                            echo "Service ID: ${RENDER_PRODUCTION_SERVICE_ID}"
                            
                            echo "Using Render API for deployment..."
                            curl -s -w "\\n%{http_code}" -X POST \
                            "https://api.render.com/v1/services/${RENDER_PRODUCTION_SERVICE_ID}/deploys" \
                            -H "Authorization: Bearer \${RENDER_API_KEY}" \
                            -H "Content-Type: application/json" \
                            -d '{
                                "clearCache": "do_not_clear", "imageUrl": "'"${env.DOCKER_IMAGE}"'"}'
                            
                            echo "Production deployment initiated successfully"
                            echo "Waiting 30 seconds for deployment to stabilize..."
                            sleep 30
                        """
                    }
                }
            }
        }

        stage('14. Health Check') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                }
            }
            steps {
                script {
                    echo '============================================'
                    echo 'CD STAGE 14: HEALTH CHECK'
                    echo '============================================'
                    
                    def healthUrl = ""
                    if (env.BRANCH_NAME == 'main') {
                        healthUrl = "${RENDER_PRODUCTION_URL}/health"
                        echo "Checking production environment: ${healthUrl}"
                    } else if (env.BRANCH_NAME == 'develop') {
                        healthUrl = "${RENDER_STAGING_URL}/health"
                        echo "Checking staging environment: ${healthUrl}"
                    }
                    
                    if (healthUrl) {
                        retry(5) {
                            sh """
                                echo "Performing health check on: ${healthUrl}"
                                
                                response=\$(curl -s -o /dev/null -w "%{http_code}" ${healthUrl})
                                
                                if [ "\$response" = "200" ]; then
                                    echo "✓ Health check passed (HTTP 200)"
                                    
                                    # Get detailed health status
                                    echo "Fetching detailed health status..."
                                    curl -s ${healthUrl} | python3 -m json.tool || curl -s ${healthUrl}
                                    
                                    exit 0
                                else
                                    echo "✗ Health check failed (HTTP \$response)"
                                    echo "Retrying in 10 seconds..."
                                    sleep 10
                                    exit 1
                                fi
                            """
                        }
                        echo "Deployment verified successfully!"
                    }
                }
            }
        }
    }

    post {
        always {
            echo '============================================'
            echo 'POST-BUILD: CLEANUP'
            echo '============================================'
            // Archive artifacts BEFORE cleanup
            echo 'Archive artifacts...'
            archiveArtifacts artifacts: 'bin/**,coverage.out,coverage.html,gosec-report.json,govulncheck-report.json,trivy-report.json,trivy-image-report.json', allowEmptyArchive: true

            echo 'Cleaning up Go cache...'
            sh '''
                chmod -R u+w .gomodcache .gocache .go 2>/dev/null || true
                rm -rf .gomodcache .gocache .go || true
            '''
        }

        success {
            echo 'CI/CD pipeline completed successfully'
            script {
                if (env.DOCKER_IMAGE) {
                    echo "Docker image built and pushed: ${env.DOCKER_IMAGE}"
                }
                def deployEnv = env.DEPLOY_ENVIRONMENT ?: 'N/A'
                def message = """
                    ✅ *BUILD SUCCESS*
                    Job: ${env.JOB_NAME} #${env.BUILD_NUMBER}
                    Branch: ${env.BRANCH_NAME}
                    Environment: ${deployEnv}
                    Duration: ${currentBuild.durationString}
                    Build URL: ${env.BUILD_URL}
                """.stripIndent()
                
                // Send Slack notifications via Slack Plugin
                slackSend(
                    channel: env.SLACK_CHANNEL,
                    color: 'good',
                    message: message,
                    tokenCredentialId: env.SLACK_BOT_TOKEN,
                    botUser: true
                )
            }
        }

        failure {
            echo 'CI/CD pipeline failed'
            script {
                if (env.DOCKER_IMAGE) {
                    echo "Docker image that failed: ${env.DOCKER_IMAGE}"
                }
                def deployEnv = env.DEPLOY_ENVIRONMENT ?: 'N/A'
                def message = """
                    ❌ *BUILD FAILED*
                    Job: ${env.JOB_NAME} #${env.BUILD_NUMBER}
                    Branch: ${env.BRANCH_NAME}
                    Environment: ${deployEnv}
                    Duration: ${currentBuild.durationString}
                    Build URL: ${env.BUILD_URL}
                """.stripIndent()
                
                // Send Slack notifications via Slack Plugin
                slackSend(
                    channel: env.SLACK_CHANNEL,
                    color: 'danger',
                    message: message,
                    tokenCredentialId: env.SLACK_BOT_TOKEN,
                    botUser: true
                )
            }
        }
        aborted {
            echo '🛑 Pipeline was aborted!'
            script {
                def message = """
                    🛑 *BUILD ABORTED*
                    Build URL: ${env.BUILD_URL}
                """.stripIndent()
                
                // Send Slack notifications via Slack Plugin
                slackSend(
                    channel: env.SLACK_CHANNEL,
                    color: 'warning',
                    message: message,
                    tokenCredentialId: env.SLACK_BOT_TOKEN,
                    botUser: true
                )
            }
        }
    }
}