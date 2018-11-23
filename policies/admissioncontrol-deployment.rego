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
    "reason": concat(", ", deny)
} {
    count(deny) > 0
}


# Admission checks

deny["No Git repository label"] {
    policies.archivingRequired
    not input.request.object.metadata.labels["git.repository"]
}

deny["Invalid Git repository label"] {
    policies.archivingRequired
    not startswith(input.request.object.metadata.labels["git.repository"], "https://")
}

deny["No Git commit hash label"] {
    policies.archivingRequired
    not input.request.object.metadata.labels["git.commit"]
}

deny["Invalid Git commit hash label"] {
    policies.archivingRequired
    not re_match(`^[a-f0-9]{40}$`, input.request.object.metadata.labels["git.commit"])
}

deny[msg] {
    endswith(input.request.object.spec.template.spec.containers[i].image, ":latest")
    msg = sprintf("No explicit image version for the container %s", [
        input.request.object.spec.template.spec.containers[i].name
    ])
}

# Admission post-processors

processors["https://archiving-processor.admissioncontrol.svc/archive-deployment"] { policies.archivingRequired }
processors["https://auditing-processor.admissioncontrol.svc/audit-deployment"] { policies.auditingRequired }
