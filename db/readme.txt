Before applying the dump, please, create (or truncate) database with name "dappctrldb" owned by the user "dappmanager". 

CREATE ROLE dappmanager WITH PASSWORD 'dappmanager';
ALTER ROLE dappmanager WITH LOGIN;

CREATE DATABASE dappctrldb WITH OWNER dappmanager;
