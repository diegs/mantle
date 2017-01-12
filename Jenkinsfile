#!groovy

properties([
    buildDiscarder(logRotator(daysToKeepStr: '20', numToKeepStr: '30')),

    [$class: 'GithubProjectProperty',
     projectUrlStr: 'https://github.com/coreos/mantle'],

    [$class: 'CopyArtifactPermissionProperty',
     projectNames: '*'],

    parameters([
        choice(name: 'GOARCH',
               choices: "amd64\narm64",
               description: 'target architecture for building binaries')
    ]),

    pipelineTriggers([pollSCM('H/15 * * * *')])
])

node('docker') {
    stage('SCM') {
        checkout scm
    }

    stage('Build') {
        sh "docker run --rm -e CGO_ENABLED=1 -e GOARCH=${params.GOARCH} -v \"\$PWD\":/usr/src/myapp -w /usr/src/myapp golang:1.7.1 ./build"
    }

    stage('Test') {
        sh 'docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.7.1 ./test'
    }

    stage('Post-build') {
        archiveArtifacts artifacts: 'bin/**', fingerprint: true, onlyIfSuccessful: true
    }
}