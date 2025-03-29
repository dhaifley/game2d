function Update(game)
	if game == nil then
		print("Global game table is nil")
		return game
	end

	if game.objects == nil then
		print("Game objects table is nil")
		return game
	end

	local obj = game.subject

	if obj == nil then
		print("Game subject not found")
		return game
	end

	if obj.data == nil then
		obj.data = {}
	end

	obj.data.dir = ""

	if obj.data.speed == nil then
		obj.data.speed = 2
	end

	if game.keys ~= nil then
		for _, k in pairs(game.keys) do
			if k == 0 or k == 29 then
				obj.data.dir = obj.data.dir .. "a"
			elseif k == 3 or k == 30 then
				obj.data.dir = obj.data.dir .. "d"
			elseif k == 22 or k == 31 then
				obj.data.dir = obj.data.dir .. "w"
			elseif k == 18 or k == 28 then
				obj.data.dir = obj.data.dir .. "s"
			elseif k >= 43 and k <= 52 then
				obj.data.speed = k - 43
			end
		end
	end

	if string.find(obj.data.dir, "a") then
		obj.x = obj.x - obj.data.speed
	end

	if string.find(obj.data.dir, "d") then
		obj.x = obj.x + obj.data.speed
	end

	if obj.x > 640 - obj.w then
		obj.x = 640 - obj.w
	elseif obj.x < 0 then
		obj.x = 0
	end

	if string.find(obj.data.dir, "w") then
		obj.y = obj.y - obj.data.speed
	end

	if string.find(obj.data.dir, "s") then
		obj.y = obj.y + obj.data.speed
	end

	if obj.y > 480 - obj.h then
		obj.y = 480 - obj.h
	elseif obj.y < 0 then
		obj.y = 0
	end

	game.subject = obj

	return game
end
