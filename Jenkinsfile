#!groovy

def IMAGE_REPO = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/pagerduty-operator"
def DOCKER_TAG = (env.BRANCH_NAME == 'main') ? 'latest' : env.GIT_COMMIT
def CI_IMAGE = "pagerduty-operator-ci:${env.BRANCH_NAME}"
def RELEASE_TAG = "${env.BRANCH_NAME}-${env.GIT_COMMIT.take(8)}"

pipeline {
    agent {
        label 'infrastructure'
    }
    environment {
        AWS_DEFAULT_REGION = "us-west-2"
        IMG = '${IMAGE_REPO}:${DOCKER_TAG}'
    }

    stages {

        stage("Build CI Image") {
            steps {
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
                    "Manifests": { sh "docker run --rm --env IMG=${IMG}  ${CI_IMAGE} output_manifests" },
                    "Deployment Image": { sh "docker build -t ${IMAGE_REPO}:${DOCKER_TAG} ." }
                )
            }
        }

        stage ('Push') {
            when {
                branch "main"
            }
            steps {
                sh "docker push ${IMG}"
                sh "DOCKER RUN --RM --env IMG=${IMG} ${CI_IMAGE} release RELEASE_TAG=${RELEASE_TAG}"
            }
        }
    }
}