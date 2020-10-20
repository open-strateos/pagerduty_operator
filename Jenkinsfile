#!groovy

def IMAGE_REPO = "742073802618.dkr.ecr.us-west-2.amazonaws.com/strateos/pagerduty-operator"
def DOCKER_TAG = (env.BRANCH_NAME == 'main') ? 'latest' : env.BRANCH_NAME

pipeline {
    agent {
        label 'infrastructure'
    }
    environment {
        AWS_DEFAULT_REGION = "us-west-2"
        IMG = IMAGE_REPO + ':' + DOCKER_TAG
    }
    options {
        skipDefaultCheckout()
    }

    stages {
        stage('Test') {
            steps {
                sh "docker built --target tester ."
            }
        }

        stage('Build') {
            parallel {
                stage('Docker') {
                    sh "make docker-build"
                }

                stage('Manifests') {
                    sh make output-manifests
                }
            }
        }

        stage ('Push') {
            parallel {
                stage ('docker push') {
                    sh "make docker-push"
                }
            }
        }
    }
}