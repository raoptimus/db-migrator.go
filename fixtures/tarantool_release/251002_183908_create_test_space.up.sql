box.schema.space.create('test', {if_not_exists = true});

box.space.test:format({{'name', type = 'string', is_nullable = false},{'rank', type = 'unsigned', is_nullable = false}});

box.space.test:create_index('primary', {parts = {'name'}, if_not_exists = true});

box.space.test:create_index('rank', {parts = {{'rank'}}, if_not_exists = true});

box.schema.space.create('test2', {if_not_exists = true});
