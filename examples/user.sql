create table user
(
  id int auto_increment primary key,
  name varchar(40) null,
  status enum("active","deleted") DEFAULT "active",
  created timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
)
  engine=InnoDB;


INSERT Into user (`id`,`name`) VALUE (1,"Jack");
UPDATE user SET name="Jonh" WHERE id=1;
DELETE FROM user WHERE id=1;