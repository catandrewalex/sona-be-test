-- use this if you're accessing the MySQL server using MariaDB
-- ALTER USER 'root'@'%' IDENTIFIED WITH mysql_native_password BY 'p4ssw0rd';
ALTER USER 'root'@'%' IDENTIFIED BY 'p4ssw0rd';
DROP USER IF EXISTS 'sonamusica'@'%';
-- use this if you're accessing the MySQL server using MariaDB
-- CREATE USER 'sonamusica'@'%' IDENTIFIED WITH mysql_native_password BY 'p4ssw0rd';
CREATE USER 'sonamusica'@'%' IDENTIFIED BY 'p4ssw0rd';
GRANT ALL PRIVILEGES ON sonamusica_administration_backend.* TO 'sonamusica'@'%';
FLUSH PRIVILEGES;

-- This is to allow any user to create trigger, function, and method
SET GLOBAL log_bin_trust_function_creators = 1;
