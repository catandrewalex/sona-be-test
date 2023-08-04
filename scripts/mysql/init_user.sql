DROP USER IF EXISTS 'sonamusica'@'localhost';
CREATE USER 'sonamusica'@'%' IDENTIFIED BY 'p4ssw0rd';
GRANT ALL PRIVILEGES ON sonamusica_administration_backend.* TO 'sonamusica'@'%';
FLUSH PRIVILEGES;
