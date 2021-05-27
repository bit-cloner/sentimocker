# Sentimocker

:star: Star me on GitHub — it motivates collaborators to contribute

sentimocker is an interactive CLI tool that can help with testing Sentinel polocies.
Sentinel is Hashicorp's policy as code language.Testing sentinel policies locally while
developing requires downloading mock terraform plan data.This data can be downloaded
from Terraform enterprise or Terraform cloud UI.It is a manual process of navigating to the right
workspace then to the right run and clicking a button which downloads a tar.gz file.
You can learn about testing Sentinel policies [here](https://medium.com/hashicorp-engineering/writing-and-testing-sentinel-policies-for-terraform-enterprise-2611112d5752)

### Why do I need this
The current approach of manually downloading mock files from Terraform UI breaks the flow of
building and testing policies. Developers need to login to terraform cloud account in a browser
then look for the download mock files button then untar the package.As developers love their IDE
and Terminal ecosystem it would improve their velocity if they could write and test
sentinel policies without shifting to other systems.

### How does sentimocker work
Sentimocker uses Hashicorp's official golang SDK for terraform enterprise to make
API calls to terraform cloud. In order to do this sentimocker needs an API token with appropriate permissions
to talk to terraform cloud.
#### Inputs & selections
`API token : A bearer token from terraform cloud account`

sentimocker interactively asks to choose from a list of options to get the right mock data
for the right plan/run.
It then unpacks the tar.gz file into /mocks directory
Developers can then copy requred files into test/ folder and edit mock files as permissionspass and
fail criteria.
