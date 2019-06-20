-- Delete rating from related_type enum.
DELETE FROM jobs WHERE related_type='rating';
DELETE FROM pg_enum WHERE enumlabel = 'rating' AND enumtypid = ( SELECT oid FROM pg_type WHERE typname = 'related_type' );

DROP TABLE ratings;
DROP TABLE closings;
DROP TYPE closing_type;
