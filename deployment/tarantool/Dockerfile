FROM tarantool/tarantool:2.6.0
COPY init.lua /opt/tarantool
COPY searcher.lua /opt/tarantool
CMD ["tarantool", "/opt/tarantool/init.lua"]