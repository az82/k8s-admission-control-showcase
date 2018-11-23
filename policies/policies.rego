package policies

import data.compliance

# Intermediate  policies derived from the compliance framework

archivingRequired = compliance.godb
auditingRequired = any([compliance.godb, compliance.hippa, compliance.bsiC5])
noSnapshots = any([compliance.godb, compliance.bsiC5, compliance.coporateSecurityGuideline])
