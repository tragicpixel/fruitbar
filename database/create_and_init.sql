BEGIN;

CREATE TABLE IF NOT EXISTS public.fruit_orders
(
    id bigint NOT NULL DEFAULT nextval('fruit_orders_id_seq'::regclass),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    num_apples bigint,
    num_bananas bigint,
    num_oranges bigint,
    num_cherries bigint,
    cash boolean,
    "number" text COLLATE pg_catalog."default",
    cardholder_name text COLLATE pg_catalog."default",
    expiration_date text COLLATE pg_catalog."default",
    zipcode text COLLATE pg_catalog."default",
    cvv text COLLATE pg_catalog."default",
    subtotal numeric,
    tax numeric,
    tip_amt numeric,
    total numeric,
    CONSTRAINT fruit_orders_pkey PRIMARY KEY (id)
)

TABLESPACE pg_default;

ALTER TABLE public.fruit_orders
    OWNER to postgres;
-- Index: idx_fruit_orders_deleted_at

-- DROP INDEX public.idx_fruit_orders_deleted_at;

CREATE INDEX idx_fruit_orders_deleted_at
    ON public.fruit_orders USING btree
    (deleted_at ASC NULLS LAST)
    TABLESPACE pg_default;

-- need to fix these values:
INSERT INTO orders (
	order_timestamp,
	pay_with_card,
	card_number,
	card_holder_name,
	card_exp_date,
	card_zipcode,
	card_cvv,
	subtotal,
	tax,
	tip,
	total,
	n_apples,
	n_bananas,
	n_cherries,
	n_oranges
)
VALUES
   (
	   CURRENT_TIMESTAMP,
	   true,
	   string_to_array('22222222',''), string_to_array('Abbot',''), 
	   string_to_array('0924',''), string_to_array('12345',''), string_to_array('111',''),
	   100.00, 20.000, 5.00, 125.00,
	   1, 3, 1, 0
   ),
   (
	   CURRENT_TIMESTAMP,
	   true,
	   string_to_array('333333',''), string_to_array('Costello',''), 
	   string_to_array('2323',''), string_to_array('54321',''), string_to_array('111',''),
	   133.00, 20.300, 5.00, 125.00,
	   1, 4, 5, 0
   ),
   (
	   CURRENT_TIMESTAMP,
	   true,
	   string_to_array('244444',''), string_to_array('Mario',''), 
	   string_to_array('1233',''), string_to_array('12555',''), string_to_array('666',''),
	   500.00, 20.000, 5.00, 625.00,
	   1, 2, 1, 99
   ),
   (
	   CURRENT_TIMESTAMP,
	   true,
	   string_to_array('22222222',''), string_to_array('Luigi',''), 
	   string_to_array('0924',''), string_to_array('12335',''), string_to_array('112',''),
	   100.00, 20.000, 5.00, 125.00,
	   25, 1, 7, 0
   ),
   (
	   CURRENT_TIMESTAMP,
	   true,
	   string_to_array('22222222',''), string_to_array('Jerry',''), 
	   string_to_array('0924',''), string_to_array('12345',''), string_to_array('111',''),
	   100.00, 20.000, 5.00, 125.00,
	   1, 2, 321, 0
   ),
   (
	   CURRENT_TIMESTAMP,
	   false,
	   string_to_array('',''), string_to_array('George',''), 
	   string_to_array('',''), string_to_array('',''), string_to_array('',''),
	   100.00, 20.000, 0.00, 120.00,
	   33, 1, 1, 6
   );

COMMIT;