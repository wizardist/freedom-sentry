# freedom-sentry-helm

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

MediaWiki bot for Wikipedia to suppress edit metadata for sensitive articles

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| wizardist |  |  |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| config.apiEndpoint | string | `""` | API endpoint for MediaWiki (required) |
| config.listName | string | `""` | Name of the suppression list page (required) |
| config.skipInitFullscan | bool | `false` | Skip initial full scan on startup |
| env | list | `[]` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"REGISTRY_URL/freedom-sentry"` | Image repository (override in production values) |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion |
| imagePullSecrets | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.fsGroup | int | `65534` |  |
| podSecurityContext.runAsNonRoot | bool | `true` |  |
| podSecurityContext.runAsUser | int | `65534` |  |
| replicaCount | int | `1` | Single replica - stateless daemon |
| resources.limits.cpu | string | `"200m"` |  |
| resources.limits.memory | string | `"128Mi"` |  |
| resources.requests.cpu | string | `"50m"` |  |
| resources.requests.memory | string | `"64Mi"` |  |
| secret.accessToken | string | `""` | Access token value (only used if create is true) |
| secret.create | bool | `true` | Create a secret containing the access token |
| secret.existingSecret | string | `""` | Name of existing secret to use (if create is false) Secret must contain key 'accessToken' |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tolerations | list | `[]` |  |

