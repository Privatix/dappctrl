-- Offering ip types.
CREATE TYPE ip_type AS ENUM ('residential','datacenter', 'mobile');

ALTER TABLE offerings
ADD ip_type ip_type NOT NULL; 
