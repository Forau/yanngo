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
  $ sed -i 's/int32/int64/g'  *.go                      # Why use 32 bit ints? Its not the year 2000
  $ gofmt -s -w *.go
  
