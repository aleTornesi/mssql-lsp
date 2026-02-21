create trigger update_timestamp BEFORE update 
on users for EACH row begin
set NEW.updated_at = NOW();

end;
