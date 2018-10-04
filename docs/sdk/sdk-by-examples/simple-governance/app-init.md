# Application Initialization 

In the root of your fork of the SDK, create an `app` and `cmd` folder. In this folder, we will create the main file for our application, `app.go` and the repository to handle REST and CLI commands for our app. 

```bash
mkdir app cmd 
mkdir -p cmd/simplegovcli cmd/simplegovd
touch app/app.go cmd/simplegovcli/main.go cmd/simplegovd/main.go
```

We will take care of these files later in the tutorial. The first step is to take care of our simple governance module.