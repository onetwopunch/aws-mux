# aws-mux

Have you ever had the problem when working with `terraform`, `docker` or some other tool that only or primarily supports AWS auth via environment variables but you only have credentials/config in ~/.aws/credentials and/or ~/.aws/config? This is especially difficult if you have multiple accounts with STS with MFA or even a GovCloud account as well as a normal account. `aws-mux` allows you to pick which profile you want and have those credentials, whether they be STS or standard API keys into the file `~/.aws/env` for you to source and use in subsequent scripts.

## Usage
```
go get https://github.com/onetwopunch/aws-mux
aws-mux && source ~/.aws/env

terraform plan
```

You might even consider using this as part of a wrapper script. As an example:

```
#!/bin/bash

aws-mux && source ~/.aws/env
docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
           -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
           -e "OUTPUT_FILE=s3://my-bucket/output.csv" scalesec/fedrampup
```
