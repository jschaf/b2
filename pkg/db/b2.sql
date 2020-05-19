-- b2.sql is the SQL commands to create the b2 sqlite database.

-- The raw results of a fetch from an external source.
create table raw_fetch
(
    url        text      not null primary key,
    fetch_time timestamp not null default current_timestamp,
    -- The content of the response. Maybe empty if the url was an asset.
    content    text      not null
);

-- The raw assets returned from a fetch like a PDF or image.
create table raw_asset
(
    url        text references raw_fetch,
    local_path text not null,
    primary key(url, local_path)
);

-- The content to show for a link preview. Not necessarily associated with a
-- raw_fetch if manually inserted.
create table link_preview
(
    url         text      not null primary key,
    update_time timestamp not null default current_timestamp,
    title       text      not null,
    body_html   text      not null
);


