FROM openjdk:8u131-jre-alpine

WORKDIR /connectors/egeria

COPY target/egeriaconnector-1.0-SNAPSHOT-jar-with-dependencies.jar target/egeriaconnector-1.0-SNAPSHOT-jar-with-dependencies.jar

RUN apk add --no-cache curl

ENTRYPOINT ["java", "-cp", "target/egeriaconnector-1.0-SNAPSHOT-jar-with-dependencies.jar:src/lib/*", "com.ibm.egeria.EgeriaConnector"]
