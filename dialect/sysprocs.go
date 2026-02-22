package dialect

// SystemProcedure represents a system stored procedure with its signature.
type SystemProcedure struct {
	Name   string
	Detail string
	Doc    string
}

// MSSQLSystemProcedures returns the commonly used MSSQL system procedures.
func MSSQLSystemProcedures() []SystemProcedure {
	return mssqlSystemProcedures
}

// MSSQLDMVs returns MSSQL dynamic management views (sys.dm_*).
func MSSQLDMVs() []SystemProcedure {
	return mssqlDMVs
}

// MSSQLSystemViews returns MSSQL system catalog views (sys.*).
func MSSQLSystemViews() []SystemProcedure {
	return mssqlSystemViews
}

var mssqlSystemProcedures = []SystemProcedure{
	// sp_ procedures
	{Name: "sp_help", Detail: "System Procedure", Doc: "sp_help [ @objname = ] 'name'\n\nReports information about a database object."},
	{Name: "sp_helptext", Detail: "System Procedure", Doc: "sp_helptext [ @objname = ] 'name'\n\nDisplays the definition of a user-defined rule, default, unencrypted T-SQL stored procedure, user-defined T-SQL function, trigger, computed column, CHECK constraint, view, or system object."},
	{Name: "sp_helpindex", Detail: "System Procedure", Doc: "sp_helpindex [ @objname = ] 'name'\n\nReports information about the indexes on a table or view."},
	{Name: "sp_helpdb", Detail: "System Procedure", Doc: "sp_helpdb [ [ @dbname = ] 'name' ]\n\nReports information about a specified database or all databases."},
	{Name: "sp_columns", Detail: "System Procedure", Doc: "sp_columns [ @table_name = ] 'name'\n\nReturns column information for the specified objects."},
	{Name: "sp_tables", Detail: "System Procedure", Doc: "sp_tables [ @table_name = ] 'name'\n\nReturns a list of objects that can appear in the FROM clause."},
	{Name: "sp_stored_procedures", Detail: "System Procedure", Doc: "sp_stored_procedures [ @sp_name = ] 'name'\n\nReturns a list of stored procedures in the current environment."},
	{Name: "sp_databases", Detail: "System Procedure", Doc: "sp_databases\n\nLists databases in the server instance."},
	{Name: "sp_who", Detail: "System Procedure", Doc: "sp_who [ @loginame = ] 'login'\n\nProvides information about current users, sessions, and processes."},
	{Name: "sp_who2", Detail: "System Procedure", Doc: "sp_who2 [ @loginame = ] 'login'\n\nProvides extended information about current users, sessions, and processes."},
	{Name: "sp_lock", Detail: "System Procedure", Doc: "sp_lock [ @spid = ] 'process_id'\n\nReports information about locks."},
	{Name: "sp_spaceused", Detail: "System Procedure", Doc: "sp_spaceused [ @objname = ] 'objname'\n\nDisplays the number of rows, disk space reserved, and disk space used."},
	{Name: "sp_rename", Detail: "System Procedure", Doc: "sp_rename [ @objname = ] 'object_name', [ @newname = ] 'new_name' [, [ @objtype = ] 'object_type' ]\n\nChanges the name of a user-created object."},
	{Name: "sp_depends", Detail: "System Procedure", Doc: "sp_depends [ @objname = ] 'object'\n\nDisplays information about database object dependencies."},
	{Name: "sp_executesql", Detail: "System Procedure", Doc: "sp_executesql [ @stmt = ] N'tsql_string' [, @params = N'@param_def' [, @param1 = value1 [...]]]\n\nExecutes a T-SQL statement or batch that can be reused many times, or one that has been built dynamically."},
	{Name: "sp_configure", Detail: "System Procedure", Doc: "sp_configure [ @configname = ] 'option_name' [, @configvalue = ] 'value'\n\nDisplays or changes global configuration settings for the current server."},
	{Name: "sp_addrolemember", Detail: "System Procedure", Doc: "sp_addrolemember [ @rolename = ] 'role', [ @membername = ] 'security_account'\n\nAdds a database user to a database role."},
	{Name: "sp_droprolemember", Detail: "System Procedure", Doc: "sp_droprolemember [ @rolename = ] 'role', [ @membername = ] 'security_account'\n\nRemoves a security account from a SQL Server role."},
	{Name: "sp_helprole", Detail: "System Procedure", Doc: "sp_helprole [ @rolename = ] 'role'\n\nReturns information about the roles in the current database."},
	{Name: "sp_helprolemember", Detail: "System Procedure", Doc: "sp_helprolemember [ @rolename = ] 'role'\n\nReturns information about the members of a role."},
	{Name: "sp_helpuser", Detail: "System Procedure", Doc: "sp_helpuser [ @name_in_db = ] 'security_account'\n\nReports information about database-level principals."},
	{Name: "sp_helpconstraint", Detail: "System Procedure", Doc: "sp_helpconstraint [ @objname = ] 'table'\n\nReturns a list of all constraint types and their properties."},
	{Name: "sp_helptrigger", Detail: "System Procedure", Doc: "sp_helptrigger [ @tabname = ] 'table'\n\nReturns the type of DML triggers defined on the specified table."},
	{Name: "sp_helpfile", Detail: "System Procedure", Doc: "sp_helpfile [ @filename = ] 'name'\n\nReturns physical names and attributes of files associated with the current database."},
	{Name: "sp_helpfilegroup", Detail: "System Procedure", Doc: "sp_helpfilegroup [ @filegroupname = ] 'name'\n\nReturns the names and attributes of filegroups associated with the current database."},
	// xp_ procedures
	{Name: "xp_cmdshell", Detail: "Extended Procedure", Doc: "xp_cmdshell 'command_string' [, NO_OUTPUT]\n\nSpawns a Windows command shell and passes in a string for execution."},
	{Name: "xp_fileexist", Detail: "Extended Procedure", Doc: "xp_fileexist 'file_name' [, @file_exists OUTPUT [, @file_is_a_directory OUTPUT [, @parent_directory_exists OUTPUT]]]\n\nDetermines whether a file exists on disk."},
	{Name: "xp_fixeddrives", Detail: "Extended Procedure", Doc: "xp_fixeddrives\n\nLists all fixed drives and the amount of free space on each drive."},
	{Name: "xp_loginconfig", Detail: "Extended Procedure", Doc: "xp_loginconfig ['config_name']\n\nReports the login security configuration."},
	{Name: "xp_msver", Detail: "Extended Procedure", Doc: "xp_msver [optname]\n\nReturns version information about SQL Server."},
	{Name: "xp_servicecontrol", Detail: "Extended Procedure", Doc: "xp_servicecontrol 'action', 'service'\n\nControls a Windows service."},
	{Name: "xp_sprintf", Detail: "Extended Procedure", Doc: "xp_sprintf @string OUTPUT, @format, @argument1 [...]\n\nFormats and stores a series of characters and values in the string output parameter."},
	{Name: "xp_sscanf", Detail: "Extended Procedure", Doc: "xp_sscanf 'string', 'format', @argument1 [...]\n\nReads data from the string into the argument locations specified by each format argument."},
	{Name: "xp_logevent", Detail: "Extended Procedure", Doc: "xp_logevent error_number, 'message' [, 'severity']\n\nLogs a user-defined message in the SQL Server log file and in the Windows Event Viewer."},
	{Name: "xp_sqlmaint", Detail: "Extended Procedure", Doc: "xp_sqlmaint 'switch_string'\n\nCalls the sqlmaint utility with a string that contains sqlmaint switches."},
}

var mssqlDMVs = []SystemProcedure{
	// Execution
	{Name: "sys.dm_exec_requests", Detail: "DMV", Doc: "Returns information about each request executing within SQL Server."},
	{Name: "sys.dm_exec_sessions", Detail: "DMV", Doc: "Returns one row per authenticated session on SQL Server."},
	{Name: "sys.dm_exec_connections", Detail: "DMV", Doc: "Returns information about the connections established to this instance of SQL Server."},
	{Name: "sys.dm_exec_query_stats", Detail: "DMV", Doc: "Returns aggregate performance statistics for cached query plans."},
	{Name: "sys.dm_exec_query_plan", Detail: "DMV", Doc: "Returns the showplan in XML format for the batch specified by the plan handle."},
	{Name: "sys.dm_exec_sql_text", Detail: "DMV", Doc: "Returns the text of the SQL batch identified by the specified sql_handle."},
	{Name: "sys.dm_exec_procedure_stats", Detail: "DMV", Doc: "Returns aggregate performance statistics for cached stored procedures."},
	{Name: "sys.dm_exec_cached_plans", Detail: "DMV", Doc: "Returns a row for each query plan cached by SQL Server."},
	{Name: "sys.dm_exec_trigger_stats", Detail: "DMV", Doc: "Returns aggregate performance statistics for cached triggers."},
	// OS
	{Name: "sys.dm_os_wait_stats", Detail: "DMV", Doc: "Returns information about all the waits encountered by threads that executed."},
	{Name: "sys.dm_os_waiting_tasks", Detail: "DMV", Doc: "Returns information about the wait queue of tasks that are waiting on some resource."},
	{Name: "sys.dm_os_performance_counters", Detail: "DMV", Doc: "Returns a row per performance counter maintained by the server."},
	{Name: "sys.dm_os_memory_clerks", Detail: "DMV", Doc: "Returns the set of all memory clerks currently active in the instance."},
	{Name: "sys.dm_os_buffer_descriptors", Detail: "DMV", Doc: "Returns information about all the data pages that are currently in the SQL Server buffer pool."},
	{Name: "sys.dm_os_schedulers", Detail: "DMV", Doc: "Returns one row per scheduler in SQL Server where each scheduler is mapped to an individual processor."},
	{Name: "sys.dm_os_threads", Detail: "DMV", Doc: "Returns a list of all SQL Server operating system threads running under the SQL Server process."},
	{Name: "sys.dm_os_sys_info", Detail: "DMV", Doc: "Returns a miscellaneous set of useful information about the computer and the resources available to SQL Server."},
	// Index
	{Name: "sys.dm_db_index_usage_stats", Detail: "DMV", Doc: "Returns counts of different types of index operations and the time each was last performed."},
	{Name: "sys.dm_db_index_physical_stats", Detail: "DMV", Doc: "Returns size and fragmentation information for the data and indexes of the specified table or view."},
	{Name: "sys.dm_db_index_operational_stats", Detail: "DMV", Doc: "Returns current low-level I/O, locking, latching, and access method activity for each partition of a table or index."},
	{Name: "sys.dm_db_missing_index_details", Detail: "DMV", Doc: "Returns detailed information about missing indexes, excluding spatial indexes."},
	{Name: "sys.dm_db_missing_index_groups", Detail: "DMV", Doc: "Returns information about missing indexes in a specific missing index group."},
	{Name: "sys.dm_db_missing_index_group_stats", Detail: "DMV", Doc: "Returns summary information about groups of missing indexes."},
	// Transaction
	{Name: "sys.dm_tran_active_transactions", Detail: "DMV", Doc: "Returns information about transactions for the instance of SQL Server."},
	{Name: "sys.dm_tran_session_transactions", Detail: "DMV", Doc: "Returns correlation information for associated transactions and sessions."},
	{Name: "sys.dm_tran_locks", Detail: "DMV", Doc: "Returns information about currently active lock manager resources."},
	{Name: "sys.dm_tran_database_transactions", Detail: "DMV", Doc: "Returns information about transactions at the database level."},
	// IO
	{Name: "sys.dm_io_virtual_file_stats", Detail: "DMV", Doc: "Returns I/O statistics for data and log files."},
}

var mssqlSystemViews = []SystemProcedure{
	{Name: "sys.objects", Detail: "System View", Doc: "Contains a row for each user-defined, schema-scoped object."},
	{Name: "sys.tables", Detail: "System View", Doc: "Returns a row for each user table in SQL Server."},
	{Name: "sys.views", Detail: "System View", Doc: "Contains a row for each view object."},
	{Name: "sys.columns", Detail: "System View", Doc: "Returns a row for each column of an object that has columns."},
	{Name: "sys.indexes", Detail: "System View", Doc: "Contains a row per index or heap of a tabular object."},
	{Name: "sys.index_columns", Detail: "System View", Doc: "Contains one row per column that is part of an index or unordered table (heap)."},
	{Name: "sys.types", Detail: "System View", Doc: "Contains a row for each system and user-defined type."},
	{Name: "sys.schemas", Detail: "System View", Doc: "Contains a row for each database schema."},
	{Name: "sys.procedures", Detail: "System View", Doc: "Contains a row for each stored procedure object."},
	{Name: "sys.parameters", Detail: "System View", Doc: "Contains a row for each parameter of an object that accepts parameters."},
	{Name: "sys.triggers", Detail: "System View", Doc: "Contains a row for each trigger object."},
	{Name: "sys.foreign_keys", Detail: "System View", Doc: "Contains a row for each object that is a FOREIGN KEY constraint."},
	{Name: "sys.foreign_key_columns", Detail: "System View", Doc: "Contains a row for each column that makes up a foreign key."},
	{Name: "sys.key_constraints", Detail: "System View", Doc: "Contains a row for each object that is a PRIMARY KEY or UNIQUE constraint."},
	{Name: "sys.check_constraints", Detail: "System View", Doc: "Contains a row for each object that is a CHECK constraint."},
	{Name: "sys.default_constraints", Detail: "System View", Doc: "Contains a row for each object that is a default definition."},
	{Name: "sys.databases", Detail: "System View", Doc: "Contains a row for each database in the instance of SQL Server."},
	{Name: "sys.server_principals", Detail: "System View", Doc: "Contains a row for every server-level principal."},
	{Name: "sys.database_principals", Detail: "System View", Doc: "Returns a row for each principal in a database."},
	{Name: "sys.database_permissions", Detail: "System View", Doc: "Returns a row for every permission or column-exception permission."},
	{Name: "sys.sql_modules", Detail: "System View", Doc: "Returns a row for each object that is an SQL language-defined module."},
	{Name: "sys.all_objects", Detail: "System View", Doc: "Shows the UNION of all schema-scoped user-defined objects and system objects."},
	{Name: "sys.all_columns", Detail: "System View", Doc: "Shows the union of all columns belonging to user-defined objects and system objects."},
	{Name: "sys.dm_sql_referenced_entities", Detail: "System View", Doc: "Returns one row for each user-defined entity referenced by name in the definition of the specified referencing entity."},
	{Name: "sys.dm_sql_referencing_entities", Detail: "System View", Doc: "Returns one row for each entity in the current database that references another user-defined entity by name."},
	{Name: "sys.stats", Detail: "System View", Doc: "Contains a row for each statistics object that exists for the tables, indexes, and indexed views."},
	{Name: "sys.partitions", Detail: "System View", Doc: "Contains a row for each partition of all the tables and most types of indexes."},
	{Name: "sys.allocation_units", Detail: "System View", Doc: "Contains a row for each allocation unit in the database."},
	{Name: "sys.filegroups", Detail: "System View", Doc: "Contains a row for each data space that is a filegroup."},
	{Name: "sys.database_files", Detail: "System View", Doc: "Contains a row per file of a database as stored in the database itself."},
	{Name: "sys.sysprocesses", Detail: "System View", Doc: "Contains information about processes running on SQL Server. Can be filtered by SPID."},
	{Name: "sys.configurations", Detail: "System View", Doc: "Contains a row for each server-wide configuration option value in the system."},
	{Name: "sys.identity_columns", Detail: "System View", Doc: "Contains a row for each column that is an identity column."},
	{Name: "sys.computed_columns", Detail: "System View", Doc: "Contains a row for each column found in sys.columns that is a computed column."},
	{Name: "sys.sequences", Detail: "System View", Doc: "Contains a row for each sequence object in a database."},
	{Name: "sys.synonyms", Detail: "System View", Doc: "Contains a row for each synonym in sys.objects."},
}
