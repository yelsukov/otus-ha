-- stored procedure to search entries by `first name` and `second name` prefixes
-- Param: prefix_first_name - prefix of first name
-- Param: prefix_second_name - prefix of second name
-- Param: offset - offset position from data starts
-- Param: limit - limit or entries in results
function search(prefix_first_name, prefix_second_name, offset, limit)
    local tuples = box.space.users.index.first_name_idx:select(prefix_first_name, { iterator = 'GE', offset = offset })
    local count = 0
    local results = {}
    for _, tuple in ipairs(tuples) do
        if count >= limit then
            return results
        end
        if string.startswith(tuple[3], prefix_first_name, 1, -1) and string.startswith(tuple[4], prefix_second_name, 1, -1) then
            table.insert(results, tuple)
            count = count + 1
        end
    end
    return results
end