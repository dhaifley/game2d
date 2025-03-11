function update(data)
	data.dir = ""

	if data.speed == nil then
		data.speed = 2
	end

	for _, k in pairs(data.keys) do
		if k == 0 or k == 29 then
			data.dir = data.dir .. "a"
		elseif k == 3 or k == 30 then
			data.dir = data.dir .. "d"
		elseif k == 22 or k == 31 then
			data.dir = data.dir .. "w"
		elseif k == 18 or k == 28 then
			data.dir = data.dir .. "s"
		elseif k >= 43 and k <= 52 then
			data.speed = k - 43
		end
	end

	if string.find(data.dir, "a") then
		data.x = data.x - data.speed
	end

	if string.find(data.dir, "d") then
		data.x = data.x + data.speed
	end

	if data.x > 400 - data.w / 2 then
		data.x = 400 - data.w / 2
	elseif data.x < -400 + data.w / 2 then
		data.x = -400 + data.w / 2
	end
	
	if string.find(data.dir, "w") then
		data.y = data.y - data.speed
	end

	if string.find(data.dir, "s") then
		data.y = data.y + data.speed
	end

	if data.y > 300 - data.h / 2 then
		data.y = 300 - data.h / 2
	elseif data.y < -300 + data.h / 2 then
		data.y = -300 + data.h / 2
	end

	return data
end
