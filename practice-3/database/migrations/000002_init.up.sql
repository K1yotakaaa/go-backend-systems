create table if not exists users (
  id serial primary key,
  name varchar(255) not null,
  email varchar(255) unique not null,
  age int not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz null,
  password_hash text null
);

create table if not exists audit_logs (
  id serial primary key,
  user_id int not null references users(id) on delete cascade,
  action varchar(64) not null,
  created_at timestamptz not null default now()
);

insert into users (name, email, age) values ('John Doe', 'john@example.com', 25);
