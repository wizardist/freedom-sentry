# freedom-sentry-helm

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.1.0](https://img.shields.io/badge/AppVersion-1.1.0-informational?style=flat-square)

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
| config.wikiDomain | string | `""` | Wiki domain (required) |
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
| resources.limits.cpu | int | `1` |  |
| resources.limits.memory | string | `"32Mi"` |  |
| resources.requests.cpu | string | `"50m"` |  |
| resources.requests.memory | string | `"16Mi"` |  |
| secret.accessToken | string | `""` | Access token value (only used if create is true) |
| secret.create | bool | `true` | Create a secret containing the access token |
| secret.existingSecret | string | `""` | Name of existing secret to use (if create is false). Secret must contain key 'accessToken' |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use |
| tolerations | list | `[]` |  |

