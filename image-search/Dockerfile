# builder stage
FROM maven:3.8.6-eclipse-temurin-17-alpine AS builder

WORKDIR /app
RUN apk add --no-cache git
RUN git clone https://github.com/alibabacloud-howto/image-search-workshop.git
WORKDIR /app/image-search-workshop
RUN mvn -B clean package -Dmaven.test.skip=true

# production stage
FROM eclipse-temurin:17-jre-alpine

WORKDIR /app
COPY --from=builder /app/image-search-workshop/target/web-image-search-engine.jar .

ENV PORT=8080
EXPOSE 8080

CMD ["java", "-jar", "web-image-search-engine.jar"]