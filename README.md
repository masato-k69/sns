# WIP
## codegen

```
api$ oapi-codegen -config ../../_context/app/oapi-codegen/config.yaml ../interface/api/openapi.yaml
api$ oapi-codegen -config ../../_context/app/oapi-codegen/v1/config.yaml ../interface/api/v1/openapi.yaml

api$ protoc -I../interface/pubsub \
    --go_out=./gen/pubsub \
    --go_opt=paths=source_relative \
    ../interface/pubsub/activity.proto
```
