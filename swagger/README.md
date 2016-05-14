# Generated files

The files here are generated with:
 $  java -jar swagger-codegen-cli.jar generate -i https://api.test.nordnet.se/next/2/api-docs/swagger -l go -o .

## Manual changes

And then manually deleted DefaultApi.go and scripts.
Added CustomDefines.go to be able to define types, or add stuff without modifying the generated files.

Rename files from CamelCase.

  $ rename 's/([A-Z])/_$1/g' *.go && rename 's/^_//g' *.go && rename 'y/A-Z/a-z/' *.go

And rename attributes ending with _.

  $ sed -i 's/Type_/Typ/g' *.go
  $ sed -i 's/Default_/IsDefault/g' *.go
  $ gofmt -s -w *.go
  