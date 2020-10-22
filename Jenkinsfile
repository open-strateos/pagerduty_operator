#!groovy


pipeline {
    agent {
        label 'infrastructure'
    }
    environment {
        AWS_DEFAULT_REGION = "us-west-2"
    }

    def IMAGE_REPO = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/pagerduty-operator"
    def GIT_COMMIT
    def DOCKER_TAG
    def CI_IMAGE
    def RELEASE_IMAGE
    def RELEASE_TAG
    def GITHUB_TOKEN = credentials('buildsecret.github_api_token')
    stages {

        stage("Define Variables") {
            steps {
                GIT_COMMIT = env.GIT_COMMIT
                DOCKER_TAG = (env.BRANCH_NAME == 'main') ? 'latest' : GIT_COMMIT
                CI_IMAGE = "pagerduty-operator-ci:${env.BRANCH_NAME}"
                RELEASE_IMAGE = "${IMAGE_REPO}:${DOCKER_TAG}"
                RELEASE_TAG = "${env.BRANCH_NAME}-${GIT_COMMIT}"
            }
        }

        stage("Build CI Image") {
            steps {
                sh "echo $GIT_COMMIT"
                sh "docker build -f Dockerfile.ci -t ${CI_IMAGE} ${WORKSPACE}"
            }
        }

        stage('Test') {
            steps {
                sh "docker run --rm ${CI_IMAGE} test"
            }
        }

        stage('Build') {
            steps {
                parallel(
                    "Manifests": { sh "docker run --rm --env IMG=${RELEASE_IMAGE}  ${CI_IMAGE} output_manifests" },
                    "Deployment Image": { sh "docker build -t ${IMAGE_REPO}:${DOCKER_TAG} ." }
                )
            }
        }

        stage ('Push') {
            when {
                branch "main"
            }
            steps {
                sh "docker push ${RELEASE_IMAGE}"
                sh "docker run --rm --env IMG=${IMG} --env GITHUB_TOKEN=${GITHUB_TOKEN} ${CI_IMAGE} release RELEASE_TAG=${RELEASE_TAG}"
            }
        }
    }
}
