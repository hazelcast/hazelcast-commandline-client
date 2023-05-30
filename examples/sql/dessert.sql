-- (c) 2023, Hazelcast, Inc. All Rights Reserved.

CREATE OR REPLACE MAPPING desserts(
    __key int,
    name varchar,
    category varchar
) TYPE IMAP OPTIONS (
    'keyFormat' = 'int',
    'valueFormat' = 'json-flat'
);

CREATE OR REPLACE MAPPING orders(
    __key int,
    dessertId int,
    itemCount int,
    dessertName varchar,
    dessertCategory varchar
) TYPE IMAP OPTIONS (
    'keyFormat' = 'int',
    'valueFormat' = 'json-flat'
);

SINK INTO desserts (__key, name, category) VALUES
(0, 'Bolo de Bolacha', 'Cakes'),
(1, 'Tiramisu', 'Cakes'),
(2, 'Cheesecake', 'Cakes'),
(3, 'Black Forest Cake', 'Cakes'),
(4, 'Red Velvet Cake', 'Cakes'),
(5, 'Victoria Sponge Cake', 'Cakes'),
(6, 'Éclair', 'Pastries'),
(7, 'Croissant', 'Pastries'),
(8, 'Baklava', 'Pastries'),
(9, 'Napoleon', 'Pastries'),
(10, 'Strudel', 'Pastries'),
(11, 'Chocolate Truffles', 'Candies'),
(12, 'Turkish Delight', 'Candies'),
(13, 'Fudge', 'Candies'),
(14, 'Toffee', 'Candies'),
(15, 'Caramel', 'Candies'),
(16, 'Chocolate Chip Cookies', 'Cookies'),
(17, 'Macarons', 'Cookies'),
(18, 'Snickerdoodle', 'Cookies'),
(19, 'Oatmeal Raisin Cookies', 'Cookies'),
(20, 'Shortbread', 'Cookies'),
(21, 'Crème Brûlée', 'Custards'),
(22, 'Flan', 'Custards'),
(23, 'Crème Caramel', 'Custards'),
(24, 'Baked Custard', 'Custards'),
(25, 'Panna Cotta', 'Custards'),
(26, 'Churros', 'Fried Desserts'),
(27, 'Funnel Cake', 'Fried Desserts'),
(28, 'Beignets', 'Fried Desserts'),
(29, 'Gulab Jamun', 'Fried Desserts'),
(30, 'Jalebi', 'Fried Desserts'),
(31, 'Banana Pudding', 'Puddings'),
(32, 'Rice Pudding', 'Puddings'),
(33, 'Bread Pudding', 'Puddings'),
(34, 'Tapioca Pudding', 'Puddings'),
(35, 'Chocolate Pudding', 'Puddings'),
(36, 'Apple Pie', 'Pies and Tarts'),
(37, 'Pecan Pie', 'Pies and Tarts'),
(38, 'Lemon Tart', 'Pies and Tarts'),
(39, 'Key Lime Pie', 'Pies and Tarts'),
(40, 'Blueberry Pie', 'Pies and Tarts'),
(41, 'Gelato', 'Frozen Desserts'),
(42, 'Sorbet', 'Frozen Desserts'),
(43, 'Sherbet', 'Frozen Desserts'),
(44, 'Ice Cream', 'Frozen Desserts'),
(45, 'Semifreddo', 'Frozen Desserts'),
(46, 'Hot Chocolate', 'Hot Beverages'),
(47, 'Mulled Wine', 'Hot Beverages'),
(48, 'Chai Tea', 'Hot Beverages'),
(49, 'Dobos Torte', 'Cakes'),
(50, 'Irish Coffee', 'Hot Beverages'),
(51, 'Sacher Torte', 'Cakes'),
(52, 'Opera Cake', 'Cakes')
;
