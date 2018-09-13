# How to update API docs

Due to the rest handlers and related data structures are distributed in many sub-folds, currently there is no tool which can automatically extract all APIs information and generate API docs. So here we have to write APIs' docs manually.

## Steps
 
1. Install the swagger-ui generate tool first.
    ```
    make update_dev_tools
    ```
2. Edit API docs
    1. Directly Edit API docs manually: `client/lcd/swaggerui/swagger.json`
    2. Edit API docs within this [SwaggerHub](https://app.swaggerhub.com). Please refer to this [document](https://app.swaggerhub.com/help/index) for how to use the about website to edit API docs.
3. Download `swagger.json` and replace the old `swagger.json` under `client/lcd/swaggerui` folds
4. Compile new gaiacli
    ```
    make install
    ```