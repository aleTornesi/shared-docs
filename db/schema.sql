
-- DROP SEQUENCE public.documents_id_seq;

CREATE SEQUENCE public.documents_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 9223372036854775807
	START 1
	CACHE 1
	NO CYCLE;
-- DROP SEQUENCE public.users_id_seq;

CREATE SEQUENCE public.users_id_seq
	INCREMENT BY 1
	MINVALUE 1
	MAXVALUE 9223372036854775807
	START 1
	CACHE 1
	NO CYCLE;-- public.users definition

-- Drop table

-- DROP TABLE public.users;

CREATE TABLE public.users (
	id int8 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE) NOT NULL,
	username varchar(50) NOT NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id),
	CONSTRAINT users_username_key UNIQUE (username)
);



-- public.documents definition

-- Drop table

-- DROP TABLE public.documents;

CREATE TABLE public.documents (
	id int8 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE) NOT NULL,
	title varchar(200) NOT NULL,
	owner_id int8 NOT NULL,
	CONSTRAINT documents_pkey PRIMARY KEY (id),
	CONSTRAINT documents_title_owner_id_key UNIQUE (title, owner_id),
	CONSTRAINT documents_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id)
);


-- public.page definition

-- Drop table

-- DROP TABLE public.page;

CREATE TABLE public.page (
	document_id int8 NOT NULL,
	page_number int2 NOT NULL,
	"content" text NOT NULL,
	CONSTRAINT page_document_id_page_number_key UNIQUE (document_id, page_number) INITIALLY DEFERRED,
	CONSTRAINT page_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id)
);


-- public.processed_events definition

CREATE TABLE public.processed_events (
	event_id varchar(36) NOT NULL,
	CONSTRAINT processed_events_pkey PRIMARY KEY (event_id)
);


-- public.document_access definition

-- Drop table

-- DROP TABLE public.document_access;

CREATE TABLE public.document_access (
	document_id int8 NOT NULL,
	user_id int8 NOT NULL,
	CONSTRAINT document_access_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id),
	CONSTRAINT document_access_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id)
);
