function Update(data)
	if game == nil then
		print("Global game table is nil")
		return
	end

	if game.objects == nil then
		print("Game objects table is nil")
		return
	end

	local obj = game.objects[data.id]
	if game.subject ~= nil and game.subject.id == data.id then
		obj = game.subject
	end

	if obj == nil then
		print("Game object not found: " .. data.id)
		local function dump(o)
			if type(o) == 'table' then
				local s = '{ '
				for k,v in pairs(o) do
					if type(k) ~= 'number' then k = '"'..k..'"' end
					s = s .. '['..k..'] = ' .. dump(v) .. ','
				end
				return s .. '} '
			else
				return tostring(o)
			end
		end

		print(dump(game.objects))
		return
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

	if obj.x > 400 - obj.w / 2 then
		obj.x = 400 - obj.w / 2
	elseif obj.x < -400 + obj.w / 2 then
		obj.x = -400 + obj.w / 2
	end

	if string.find(obj.data.dir, "w") then
		obj.y = obj.y - obj.data.speed
	end

	if string.find(obj.data.dir, "s") then
		obj.y = obj.y + obj.data.speed
	end

	if obj.y > 300 - obj.h / 2 then
		obj.y = 300 - obj.h / 2
	elseif obj.y < -300 + obj.h / 2 then
		obj.y = -300 + obj.h / 2
	end

	if game.subject ~= nil and game.subject.id == data.id then
		game.subject = obj
	else
		game.objects[data.id] = obj
	end
end
