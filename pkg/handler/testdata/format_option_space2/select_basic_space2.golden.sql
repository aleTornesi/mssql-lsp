select
  a,
  b as bb,
  c
from
  tbl
join (
  select
    a * 2 as a
  from
    new_table
) other
  on tbl.a = other.a
where
  c = 1
  and b between 3
  and 4
  or d = 'blue'