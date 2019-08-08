
def DOCKER_IMAGE_AMD64
def BBTEST_IMAGE

def dockerOptions() {
    String options = "--pull "
    options += "--label 'org.opencontainers.image.source=${env.GIT_URL}' "
    options += "--label 'org.opencontainers.image.created=${env.RFC3339_DATETIME}' "
    options += "--label 'org.opencontainers.image.revision=${env.GIT_COMMIT}' "
    options += "--label 'org.opencontainers.image.licenses=${env.LICENSE}' "
    options += "--label 'org.opencontainers.image.authors=${env.PROJECT_AUTHOR}' "
    options += "--label 'org.opencontainers.image.title=${env.PROJECT_NAME}' "
    options += "--label 'org.opencontainers.image.description=${env.PROJECT_DESCRIPTION}' "
    options += "."
    return options
}

def bbtestOptions() {
    String options = ""
    options += "-e IMAGE_VERSION=${env.GIT_COMMIT} "
    options += "-e UNIT_VERSION=${env.VERSION_MAIN}+${env.VERSION_META} "
    options += "-e UNIT_ARCH=amd64 "
    options += "-e NO_TTY=1 "
    options += "-v ${HOME}@tmp:/tmp "
    options += "-v ${HOME}/reports:/tmp/reports "
    options += "-v ${HOME}:${HOME} "
    options += "-v /var/run/docker.sock:/var/run/docker.sock:rw "
    options += "-v /var/lib/docker/containers:/var/lib/docker/containers:rw "
    options += "-v /sys/fs/cgroup:/sys/fs/cgroup:ro "
    options += "-v /run:/run:rw "
    options += "-v /run/lock:/run/lock:rw "
    return options
}

pipeline {

    agent {
        label 'master'
    }

    options {
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
        disableConcurrentBuilds()
        disableResume()
        timeout(time: 10, unit: 'MINUTES')
        timestamps()
    }

    stages {

        stage('Setup') {
            steps {
                script {
                    BBTEST_IMAGE = docker.image('jancajthaml/bbtest:amd64')

                    env.RFC3339_DATETIME = sh(
                        script: 'date --rfc-3339=ns',
                        returnStdout: true
                    ).trim()

                    env.VERSION_MAIN = sh(
                        script: 'git fetch --tags --force 2> /dev/null; tags=\$(git tag --sort=-v:refname | head -1) && ([ -z \${tags} ] && echo v0.0.0 || echo \${tags})',
                        returnStdout: true
                    ).trim() - 'v'

                    env.VERSION_META = sh(
                        script: 'git rev-parse --abbrev-ref HEAD 2> /dev/null | sed \'s:.*/::\'',
                        returnStdout: true
                    ).trim()

                    env.LICENSE = "Apache-2.0"
                    env.PROJECT_NAME = "openbank bondster-bco"
                    env.PROJECT_DESCRIPTION = "OpenBanking bondster-bco service"
                    env.PROJECT_AUTHOR = "Jan Cajthaml <jan.cajthaml@gmail.com>"
                    env.HOME = "${WORKSPACE}"
                    env.GOPATH = "${WORKSPACE}/go"
                    env.XDG_CACHE_HOME = "${env.GOPATH}/.cache"
                    env.PROJECT_PATH = "${env.GOPATH}/src/github.com/jancajthaml-openbank/bondster-bco"

                    sh """
                        mkdir -p \
                            ${env.GOPATH}/src/github.com/jancajthaml-openbank && \
                        mv \
                            ${WORKSPACE}/services/bondster-bco-rest \
                            ${env.GOPATH}/src/github.com/jancajthaml-openbank/bondster-bco-rest
                        mv \
                            ${WORKSPACE}/services/bondster-bco-import \
                            ${env.GOPATH}/src/github.com/jancajthaml-openbank/bondster-bco-import
                    """
                }
            }
        }

        stage('Fetch Dependencies') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args '--tty'
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/sync \
                        --pkg bondster-bco-rest
                    """
                }
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/sync \
                        --pkg bondster-bco-import
                    """
                }
            }
        }

        stage('Quality Gate') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args '--tty'
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/lint \
                        --pkg bondster-bco-rest
                    """
                    sh """
                        ${HOME}/dev/lifecycle/lint \
                        --pkg bondster-bco-import
                    """
                }
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/sec \
                        --pkg bondster-bco-rest
                    """
                    sh """
                        ${HOME}/dev/lifecycle/sec \
                        --pkg bondster-bco-import
                    """
                }
            }
        }

        stage('Unit Test') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args '--tty'
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/test \
                        --pkg bondster-bco-rest \
                        --output ${HOME}/reports
                    """
                }
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/test \
                        --pkg bondster-bco-import \
                        --output ${HOME}/reports
                    """
                }
            }
        }

        stage('Package') {
            agent {
                docker {
                    image 'jancajthaml/go:latest'
                    args '--tty'
                    reuseNode true
                }
            }
            steps {
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/package \
                        --pkg bondster-bco-rest \
                        --arch linux/amd64 \
                        --output ${HOME}/packaging/bin
                    """
                }
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/package \
                        --pkg bondster-bco-import \
                        --arch linux/amd64 \
                        --output ${HOME}/packaging/bin
                    """
                }
                dir(env.PROJECT_PATH) {
                    sh """
                        ${HOME}/dev/lifecycle/debian \
                        --version ${env.VERSION_MAIN}+${env.VERSION_META} \
                        --arch amd64 \
                        --source ${HOME}/packaging
                    """
                }
            }
        }

        stage('Package Docker') {
            steps {
                script {
                    DOCKER_IMAGE_AMD64 = docker.build("openbank/bondster-bco:${env.GIT_COMMIT}", dockerOptions())
                }
            }
        }

        stage('BlackBox Test') {
            steps {
                script {
                    BBTEST_IMAGE.withRun(bbtestOptions()) { c ->
                        sh """
                            docker exec -t ${c.id} \
                            python3 \
                            ${HOME}/bbtest/main.py
                        """
                    }
                }
            }
        }

        stage('Publish') {
            steps {
                script {
                    docker.withRegistry('https://registry.hub.docker.com', 'docker-hub-credentials') {
                        DOCKER_IMAGE_AMD64.push("amd64-${env.VERSION_MAIN}-${env.VERSION_META}", true)
                    }
                }
            }
        }
    }

    post {
        always {
            script {
                sh "docker rmi -f registry.hub.docker.com/openbank/bondster-bco:amd64-${env.VERSION_MAIN}-${env.VERSION_META} || :"
                sh "docker rmi -f bondster-bco:amd64-${env.GIT_COMMIT} || :"
                sh """
                    docker images \
                        --no-trunc \
                        --format '{{.ID}} {{.Tag}} {{.CreatedSince}}' | \
                    grep '<none>' | \
                    grep 'hours\\|days\\|weeks\\|months' | \
                    awk '{ print \$1 }' | \
                    xargs --no-run-if-empty docker rmi -f
                    """
                sh "docker system prune"
            }
            script {
                dir('reports') {
                    archiveArtifacts(
                        allowEmptyArchive: true,
                        artifacts: 'blackbox-tests/**/*'
                    )
                }
                dir('packaging/bin') {
                    archiveArtifacts(
                        allowEmptyArchive: true,
                        artifacts: '*'
                    )
                }

                publishHTML(target: [
                    allowMissing: true,
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: 'reports/unit-tests',
                    reportFiles: 'bondster-bco-rest-coverage.html',
                    reportName: 'Bondster BCO Rest | Unit Test Coverage'
                ])
                publishHTML(target: [
                    allowMissing: true,
                    alwaysLinkToLastBuild: false,
                    keepAll: true,
                    reportDir: 'reports/unit-tests',
                    reportFiles: 'bondster-bco-import-coverage.html',
                    reportName: 'Bondster BCO Import | Unit Test Coverage'
                ])
                junit(
                    allowEmptyResults: true,
                    testResults: 'reports/unit-tests/bondster-bco-rest-results.xml'
                )
                junit(
                    allowEmptyResults: true,
                    testResults: 'reports/unit-tests/bondster-bco-import-results.xml'
                )
                cucumber(
                    allowEmptyResults: true,
                    fileIncludePattern: '*',
                    jsonReportDirectory: 'reports/blackbox-tests/cucumber'
                )
            }
            cleanWs()
        }
        success {
            echo 'Success'
        }
        failure {
            echo 'Failure'
        }
    }
}
