protoc --proto_path=api --go_out=global_models\grpc --go_opt=paths=source_relative --go-grpc_out=global_models\grpc --go-grpc_opt=paths=source_relative api\bot\bot.proto

команда для генерации новых файлов из bot.proto
