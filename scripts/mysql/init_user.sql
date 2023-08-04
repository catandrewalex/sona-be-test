ALTER USER 'root'@'%' IDENTIFIED WITH mysql_native_password BY 'p4ssw0rd';
DROP USER IF EXISTS 'sonamusica'@'%';
CREATE USER 'sonamusica'@'%' IDENTIFIED WITH mysql_native_password BY 'p4ssw0rd';
GRANT ALL PRIVILEGES ON sonamusica_administration_backend.* TO 'sonamusica'@'%';
FLUSH PRIVILEGES;
