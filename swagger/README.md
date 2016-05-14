# Generated files

The files here are generated with:
 $  java -jar swagger-codegen-cli.jar generate -i https://api.test.nordnet.se/next/2/api-docs/swagger -l go -o .

And then manually deleted DefaultApi.go and scripts.
Added CustomDefines.go to be able to define types, or add stuff without modifying the generated files.
