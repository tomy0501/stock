INSERT INTO stock_room (facility_id, room_type_id, stock_date, stock_quantity, reservation_quantity, stop_sale_quantity) 
     WITH RECURSIVE seq AS (
         SELECT 1 AS n
         UNION ALL
         SELECT n + 1
         FROM seq
         WHERE n < 100
     )
     SELECT 0x11F179CE2CF1844482674AA19E8DB8EA AS facility_id
            ,0x11F17C2181BEF633AC70B6942CFD58C8 AS room_typeid
               ,DATE_ADD('2026-09-01', INTERVAL n DAY) AS stock_date
               ,FLOOR(RAND() * 10) AS stock_quantity
               ,FLOOR(RAND() * 5) AS reservation_quantity
               ,FLOOR(RAND() * 3) AS stop_sale_quantity
     FROM seq;