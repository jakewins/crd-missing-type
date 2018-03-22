This is a simple clone of the opkit sample project.
The [Controller](./controller.go) panics if an Added resource is missing TypeMeta.

## How to test

    # Compile
    CGO_ENABLED=0 GOOS=linux go build
    
    # Register CRD and run controller watch loop
    ./crd-missing-type &
    
    # Create a new resource..
    k apply -f sample-resource.yaml
    
    # Expect log message saying the controller saw the TypeMeta