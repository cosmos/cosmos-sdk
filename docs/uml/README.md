These diagrams are included in the docs.
They can be built using [PlantUML](plantuml.com).
The diagrams are expected to be updated only rarely, so instead of requiring PlantUML as a dependency, the diagrams are built and committed when necessary.

# How to build the diagrams

There are many ways to run PlantUML, listed [here](https://plantuml.com/running).

These diagrams are built by downloading [the `.jar` file for PlantUML version 1.2021.4](https://sourceforge.net/projects/plantuml/files/plantuml.1.2021.4.jar/download), then running the following command:

```
java -jar plantuml.1.2021.4.jar -tsvg docs/uml/*.puml
```
