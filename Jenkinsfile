pipeline {
  agent any

  options {
    timeout(time: 10, unit: 'MINUTES')
      ansiColor('xterm')
  }

  stages {
    stage('Build') {
      steps {
        dir("src/github.com/halkeye/discord-twitch-streamers") {
          checkout scm
          sh """
            export GOPATH=$WORKSPACE
            docker build -t halkeye/discord-twitch-streamers .
          """
        }
      }
    }
    stage('Deploy') {
      when { branch 'master' }
      // when { buildingTag() }
      environment {
        DOCKER = credentials('dockerhub-halkeye')
      }
      steps {
        sh 'docker login --username="$DOCKER_USR" --password="$DOCKER_PSW"'
        sh 'docker push halkeye/discord-twitch-streamers'
      }
    }
  }
  post {
    failure {
      emailext(
        attachLog: true,
        recipientProviders: [developers()],
        body: "Build failed (see ${env.BUILD_URL})",
        subject: "[JENKINS] ${env.JOB_NAME} failed",
      )
    }
  }
}
