pipeline {
    agent any

    environment {
        GO_VERSION = '1.24.1'
        GOPATH = "${env.WORKSPACE}/.go"
        GOMODCACHE = "${env.WORKSPACE}/.gomodcache"
        GOCACHE = "${env.WORKSPACE}/.gocache"
        GOTOOLCHAIN = 'auto'
        CGO_ENABLED = '0'
        GOOS = 'linux'
        GOARCH = 'amd64'
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

        stage('4. Lint') {
            steps {
                echo '============================================'
                echo 'CI STAGE 4: CODE QUALITY CHECK'
                echo '============================================'
                echo 'Running code quality checks with golangci-lint...'
                sh '''
                    which golangci-lint > /dev/null || (echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
                    golangci-lint run ./...
                '''
            }
        }

        stage('5. Build') {
            steps {
                echo '============================================'
                echo 'CI STAGE 5: BUILD APPLICATION'
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

        stage('7. Integration Test') {
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

        stage('8. Security Scan') {
            steps {
                echo '============================================'
                echo 'CI STAGE 8: SECURITY VULNERABILITY SCAN'
                echo '============================================'

                echo 'Running security scan with gosec...'
                sh '''
                    export PATH="${GOPATH}/bin:${PATH}"
                    which gosec > /dev/null || (echo "Installing gosec..."; go install github.com/securego/gosec/v2/cmd/gosec@latest)
                    gosec -fmt=json -out=gosec-report.json -exclude-dir=.gomodcache -exclude-dir=.go -exclude-dir=.gocache ./... || true
                    echo "Gosec scan completed"
                '''
                echo 'Running vulnerability scan with trivy...'
                sh '''
                    export PATH="${WORKSPACE}/bin:${PATH}"
                    which trivy > /dev/null || (echo "Installing trivy..."; curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b ${WORKSPACE}/bin)
                    trivy fs --severity HIGH,CRITICAL --format json --output trivy-report.json --skip-dirs .gomodcache . || true
                    echo "Trivy scan completed"
                '''
            }
        }

        // ===============================
        // CD STAGES
        // ===============================
    }

    post {
        always {
            echo 'Cleaning up Go cache...'
            sh '''
                chmod -R u+w .gomodcache .gocache .go 2>/dev/null || true
                rm -rf .gomodcache .gocache .go || true
            '''
        }

        success {
            echo 'CI pipeline completed successfully'
            archiveArtifacts artifacts: 'bin/**,coverage.out,coverage.html,gosec-report.json,trivy-report.json,trivy-image-report.json', allowEmptyArchive: true
        }

        failure {
            echo 'CI pipeline failed'
            archiveArtifacts artifacts: 'coverage.out,gosec-report.json,trivy-report.json,trivy-image-report.json', allowEmptyArchive: true
        }
    }
}