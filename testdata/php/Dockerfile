# PHP app on a fat base, no cache mount.
FROM php:8.3
WORKDIR /app
COPY . .
RUN composer install
CMD ["php", "-S", "0.0.0.0:8080"]
