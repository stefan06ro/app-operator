// +build k8srequired

package templates

// AppOperatorValues values required by app-operator-chart.
const AppOperatorValues = `
Installation:
  V1:
    Registry:
      Domain: quay.io
`
