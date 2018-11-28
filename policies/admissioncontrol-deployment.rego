package admissioncontrol.deployment

import data.policies


# Admission granted result
policy = {
    "allow": true,
    "processors": processors
} {
    count(deny) == 0
}

# Admission denied result
policy = {
    "allow": false,
    "reason": deny
} {
    count(deny) > 0
}


# Admission checks

deny["No Git repository annotation"] {
    policies.archivingRequired
    not gitRepositoryAnnotation(input)
}

deny["Invalid Git repository annotation"] {
    policies.archivingRequired
    not startswith(gitRepositoryAnnotation(input), "https://")
}

deny["No Git commit hash annotation"] {
    policies.archivingRequired
    not gitCommitAnnotation(input)
}

deny["Invalid Git commit hash annotation"] {
    policies.archivingRequired
    not re_match(`^[a-f0-9]{40}$`, gitCommitAnnotation(input))
}

deny[msg] {
    endswith(input.request.object.spec.template.spec.containers[i].image, ":latest")
    msg = sprintf("No explicit image version for the container %s", [input.request.object.spec.template.spec.containers[i].name])
}

# Admission post-processors

processors["https://archiving-processor.admissioncontrol.svc/archive-deployment"] { policies.archivingRequired }
processors["https://auditing-processor.admissioncontrol.svc/audit-deployment"] { policies.auditingRequired }


# Shortcut functions

gitRepositoryAnnotation(x)  = x.request.object.metadata.annotations["git.repository"]
gitCommitAnnotation(x) = x.request.object.metadata.annotations["git.commit"]
