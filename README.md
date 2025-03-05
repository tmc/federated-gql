# Federated GraphQL
A proof of concept for federating gRPC services with GraphQL federations.

## Code Generation
This project uses protoc plugins to generate code from Protocol Buffer definitions.

### protoc-gen-graphql
The `protoc-gen-graphql` tool generates GraphQL schema files from Protocol Buffer service definitions.

#### Custom Templates
You can use a custom template file with the `protoc-gen-graphql` generator by configuring the `template_path` option in your `buf.gen.yaml` file:

```yaml
- local: protoc-gen-graphql
  out: ../gen/graphql
  opt:
    - paths=source_relative
    - template_path=/path/to/your/custom/template.tmpl
```

If the specified template file is not found, the generator will fall back to using the embedded default template.
