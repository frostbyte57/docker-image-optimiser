# Single-stage Maven build: JDK + build tooling end up in the runtime image.
FROM maven:3.9-eclipse-temurin-21
WORKDIR /app
COPY . .
RUN mvn -B package
CMD ["java", "-jar", "target/app.jar"]
