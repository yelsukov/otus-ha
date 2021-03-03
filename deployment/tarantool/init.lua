box.cfg {
    listen = 3301,
    log_level = 2,
    net_msg_max = 7680,
    memtx_memory = 536870912
}

box.schema.space.create('users', { if_not_exists = true }):format({
    { name = 'id', type = 'unsigned' },
    { name = 'username', type = 'string' },
    { name = 'first_name', type = 'string' },
    { name = 'last_name', type = 'string' },
    { name = 'age', type = 'unsigned' },
    { name = 'gender', type = 'string' },
    { name = 'city', type = 'string' },
    { name = 'password', type = 'string' },
    { name = 'created_at', type = 'string' },
    { name = 'interests', type = 'string' },
})
box.space.users:create_index('primary', { type = "hash", unique = true, parts = { 'id' }, if_not_exists = true })
box.space.users:create_index('first_name_idx', { type = 'TREE', unique = false, parts = { 'first_name' }, if_not_exists = true })
box.schema.user.grant('spiderman', 'read,write,execute', 'space', 'users')

dofile('searcher.lua')