USE nutri;

DROP TABLE IF EXISTS products;
Create TABLE products (
  id int(6) unsigned NOT NULL AUTO_INCREMENT,
  title varchar(60) UNIQUE NOT NULL,
  calories int(6) NOT NULL,
  carbs int (6) NOT NULL ,
  proteins int(6) NOT NULL,
  size varchar (10) NOT NULL,
  primary key (id)
);

DROP TABLE IF EXISTS recipes;
CREATE TABLE recipes (
  id int(6) unsigned NOT NULL AUTO_INCREMENT,
  title varchar(100) UNIQUE NOT NULL,
  category int(1) NOT NULL,
  time int(15) NOT NULL,
  image varchar(300) NOT NULL,
  instructions text NOT NULL,
  calories int(6) DEFAULT 0,
  carbs int (6) DEFAULT 0 ,
  proteins int(6) DEFAULT 0,
  primary key (id)
);

DROP TABLE IF EXISTS ingredients;
CREATE TABLE ingredients (
  id int(6) unsigned NOT NULL AUTO_INCREMENT,
  recipeId int(6) unsigned NOT NULL,
  productId int(6) unsigned NOT NULL,
  value decimal(6, 4),
  index (recipeId),
  index (productId),
  primary key (id),
  foreign key (recipeId) references recipes(id),
  foreign key (productId) references products(id)
);

DROP TABLE IF EXISTS admins;
Create TABLE admins (
  id int(6) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(30) UNIQUE NOT NULL,
  pass varchar(64) NOT NULL,
  primary key (id)
);
