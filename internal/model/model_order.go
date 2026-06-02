// 【二期·订单/支付】本文件预留，一期禁止实现支付与订单逻辑。
//
// 约定：
//   - 新增 Order、Payment、Refund 等表仅写在本文件（或 model_order_*.go）
//   - 勿改动 model_shared.go 中 Article 等内容主表
//   - 权限继续通过 download_files.owner_type / owner_id / access_type 接入
//   - migrate.go 的 AutoMigrate 在小程序模型之后追加二期模型
package model
