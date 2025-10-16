box.space.test:alter({format = {{'name', type = 'string', is_nullable = false}, {'rank', type = 'unsigned', is_nullable = false}, {'description', type = 'string'}}});

box.space.test:insert{'abc', 1, 'hello world'};
