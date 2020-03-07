#!groovy
// Copyright (2020) Cobalt Speech and Language Inc.

// Keep only 10 builds on Jenkins
properties([
    buildDiscarder(logRotator(
        artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10'))
])

// setBuildStatus tells github the status of our build for a given "context"
def setBuildStatus(String context, String state, String message) {
    step([$class: "GitHubCommitStatusSetter",
          reposSource: [$class: "ManuallyEnteredRepositorySource", url: "https://github.com/cobaltspeech/log"],
          contextSource: [$class: "ManuallyEnteredCommitContextSource", context: context],
          errorHandlers: [[$class: "ChangingBuildStatusErrorHandler", result: "UNSTABLE"]],
          statusResultSource: [$class: "ConditionalStatusResultSource", results: [[$class: 'AnyBuildResult', message: message, state: state]]]])
}

def commentOnPullRequest(def text_pr) {
    json_data = sh( script: "jq -n --arg b '$text_pr' '{\"body\": \$b}'", returnStdout: true)
    def repository_url = scm.userRemoteConfigs[0].url
    def repository_name = repository_url.replace("https://github.com/","").replace(".git","")


    withCredentials([usernamePassword(credentialsId: 'jenkins-github-token',
                                      passwordVariable: 'TOKEN',
                                      usernameVariable: 'USER')]) {
	sh "curl -H 'Content-Type: application/json' --user ${USER}:${TOKEN} -s -X POST -d '${json_data}' \"https://api.github.com/repos/${repository_name}/issues/${env.CHANGE_ID}/comments\""
    }
}

if (env.CHANGE_ID || env.TAG_NAME) {
    // building a PR or a tag or a debug build
    node {
        try {
            timeout(time: 10, unit: 'MINUTES') {
                docker.image("golang:1.14").inside('-u root') {
		    sh "apt update && apt install -y jq"
		    checkout scm
		    try {

			stage("fmt-check") {
			    try {
				setBuildStatus("fmt", "PENDING", "")
				echo "Running gofmt to check coding style"
				sh "make fmt"
				setBuildStatus("fmt", "SUCCESS", "Files are correctly formatted.")
			    } catch (err) {
				setBuildStatus("fmt", "ERROR", "Some go files not correctly formatted.")
				throw err
			    }
			}

			stage("lint-check") {
			    try {
				setBuildStatus("lint", "PENDING", "")
				echo "Running linter to catch lint"
				sh "make lint"
				setBuildStatus("lint", "SUCCESS", "No linter errors")
			    } catch (err) {
				setBuildStatus("lint", "ERROR", "Some go files have linter errors.")
				throw err
			    }
			}

			stage("test") {
			    try {
				setBuildStatus("test", "PENDING", "")
				echo "Running test on all packages"
				sh "set -o pipefail; make test 2>&1 | tee test.log"

				if (env.CHANGE_ID) {
				    // send the test output to PR comment
				    def testlog = readFile "test.log"
				    msg = "# Test Report: \n```\n${testlog}\n```\n"
				    commentOnPullRequest(msg)
				}

				setBuildStatus("test", "SUCCESS", "All tests passed.")
			    } catch (err) {
				setBuildStatus("test", "ERROR", "Tests did not all pass.")
				throw err
			    }
			}
		    } finally {
			// Change ownership of everything so it can be cleaned up later
			sh "chown -R 1000:1000 ."
		    }
		}
	    }
	    mattermostSend channel: 'g-ci-notifications', color: 'good', message: "Build Successful - ${env.JOB_NAME} ${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"
	} catch (err) {
            mattermostSend channel: 'g-ci-notifications', color: 'danger', message: "Build Failed - ${env.JOB_NAME} ${env.BUILD_NUMBER} (<${env.BUILD_URL}|Open>)"
            throw err
	} finally {
            deleteDir()
	}
    }
}
