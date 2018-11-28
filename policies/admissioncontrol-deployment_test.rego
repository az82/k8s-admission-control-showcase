package admissioncontrol.deployment

test_admission_granted {
    policy == {
        "allow": true,
        "processors": {
            "https://archiving-processor.admissioncontrol.svc/archive-deployment",
            "https://auditing-processor.admissioncontrol.svc/audit-deployment"
        }
    }
    with input as {
        "request" : {
            "object" : {
                "metadata" : {
                    "annotations": {
                        "git.repository": "https://some-git-repo",
                        "git.commit": "f95f0966bbc36e95da4e55090dd2218182ce933a"
                    }
                },
                "spec": {
                    "template": {
                        "spec": {
                            "containers": [
                                {
                                    "name": "some-container",
                                    "image": "some/image:1.0"
                                }
                            ]                            
                        }
                    }
                }
            }
        }
    }
}


test_latest_tag {
    policy == {
        "allow": false,
        "reason": {"No explicit image version for the container some-container"}
    }
    with input as {
        "request" : {
            "object" : {
                "metadata" : {
                    "annotations": {
                        "git.repository": "https://some-git-repo",
                        "git.commit": "f95f0966bbc36e95da4e55090dd2218182ce933a"
                    }
                },
                "spec": {
                    "template": {
                        "spec": {
                            "containers": [
                                {
                                    "name": "some-container",
                                    "image": "some/image:latest"
                                }
                            ]                            
                        }
                    }
                }
            }
        }
    }
}


test_no_git_repo_admission {
     policy == {
        "allow": false,
        "reason": {"No Git repository annotation"}
    }
    with input as {
        "request" : {
            "object" : {
                "metadata" : {
                    "annotations": {
                        "git.commit": "f95f0966bbc36e95da4e55090dd2218182ce933a"
                    }
                },
                "spec": {
                    "template": {
                        "spec": {
                            "containers": [
                                {
                                    "name": "some-container",
                                    "image": "some/image:1.0"
                                }
                            ]                            
                        }
                    }
                }
            }
        }
    }
}


test_invalid_git_repo_admission_denied {
     policy == {
        "allow": false,
        "reason": {"Invalid Git repository annotation"}
    }
    with input as {
        "request" : {
            "object" : {
                "metadata" : {
                    "annotations": {
                        "git.repository": "git://some-repo",
                        "git.commit": "f95f0966bbc36e95da4e55090dd2218182ce933a"
                    }
                },
                "spec": {
                    "template": {
                        "spec": {
                            "containers": [
                                {
                                    "name": "some-container",
                                    "image": "some/image:1.0"
                                }
                            ]                            
                        }
                    }
                }
            }
        }
    }
}

test_no_git_commit_hash {
     policy == {
        "allow": false,
        "reason": {"No Git commit hash annotation"}
    }
    with input as {
        "request" : {
            "object" : {
                "metadata" : {
                    "annotations": {
                        "git.repository": "https://some-git-repo",
                    }
                },
                "spec": {
                    "template": {
                        "spec": {
                            "containers": [
                                {
                                    "name": "some-container",
                                    "image": "some/image:1.0"
                                }
                            ]                            
                        }
                    }
                }
            }
        }
    }
}

invalid_git_commit_hash {
     policy == {
        "allow": false,
        "reason": {"Invalid Git commit hash annotation"}
    }
    with input as {
        "request" : {
            "object" : {
                "metadata" : {
                    "annotations": {
                        "git.repository": "https://some-git-repo",
                        "git.commit": "f95f0966bbc36e95da4e55090dd2218182ce933a"
                    }
                },
                "spec": {
                    "template": {
                        "spec": {
                            "containers": [
                                {
                                    "name": "some-container",
                                    "image": "some/image:1.0"
                                }
                            ]                            
                        }
                    }
                }
            }
        }
    }
}