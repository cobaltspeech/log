#!groovy
// Copyright (2020) Cobalt Speech and Language Inc.

// Keep only 10 builds on Jenkins
properties([
    buildDiscarder(logRotator(
        artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '', numToKeepStr: '10'))
])

// build using Jenkins' shared library function
golangStdBuild()
