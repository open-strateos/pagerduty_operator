#!groovy

def IMAGE_REPO = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/pagerduty-operator"
def DOCKER_TAG = (env.BRANCH_NAME == 'main') ? 'latest' : env.GIT_COMMIT

pipeline {
    agent {
        label 'infrastructure'
    }
    environment {
        AWS_DEFAULT_REGION = "us-west-2"
        IMG = '${IMAGE_REPO}:${DOCKER_TAG}'
    }
    options {
        skipDefaultCheckout()
    }

    stages {
        stage('Test') {
            steps {
                sh "ls -la ${WORKSPACE}"
                sh "docker build --target tester ${WORKSPACE}"
            }
        }

        stage('Build') {
            steps {
                parallel(
                    "Docker": { sh "make docker-build" },
                    "Manifests": { sh "make output-manifests" }
                )
            }
        }

        stage ('Push') {
            steps {
                sh "make docker-push"
            }
        }
    }
}