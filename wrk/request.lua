local charset = {}  do -- [a-z]
    for c = 97, 122 do table.insert(charset, string.char(c)) end
end

local function randomString(length)
    if not length or length <= 0 then return '' end
    math.randomseed(os.clock()^5)
    return randomString(length - 1) .. charset[math.random(1, #charset)]
end

request = function()
  local lastName = randomString(2)
  local firstName = randomString(2)
  
  path = "/v1/users?lastName=" .. lastName .. "&firstName=" .. firstName .. "&limit=3"
  return wrk.format("GET", path)
end
