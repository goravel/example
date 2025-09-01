package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250828065556CreateSysUserTable struct{}

// Signature The unique signature for the migration.
func (r *M20250828065556CreateSysUserTable) Signature() string {
	return "20250828065556_create_sys_user_table"
}

// Up Run the migrations.
func (r *M20250828065556CreateSysUserTable) Up() error {
	if !facades.Schema().HasTable("sys_user") {
		return facades.Schema().Create("sys_user", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.String("username", 100).Comment("用户名")
			table.String("password", 255).Nullable().Comment("密码")
			table.String("nickname", 100).Comment("昵称")
			table.UnsignedBigInteger("organizationId").Comment("机构ID")
			table.String("avatar", 255).Nullable().Comment("头像")
			table.UnsignedTinyInteger("sex").Default(0).Comment("性别：0=男,1=女")
			table.String("phone", 50).Nullable().Comment("手机号")
			table.String("email", 100).Nullable().Comment("邮箱")
			table.String("card", 100).Nullable().Comment("证件号")
			table.Timestamp("birthday").Nullable().Comment("生日")
			table.String("introduction", 100).Nullable().Comment("介绍")
			table.UnsignedTinyInteger("estate").Default(0).Comment("状态：0=启用,1=锁定")
			table.Timestamp("createTime").Comment("创建时间")
			table.Timestamp("updateTime").Comment("更新时间")
			table.UnsignedTinyInteger("inside").Default(0).Comment("是否是内部用户：0=系统用户,1=内置用户")

			// 创建唯一索引
			table.Unique("username")

			// 创建普通索引
			table.Index("organizationId")
			table.Index("inside")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20250828065556CreateSysUserTable) Down() error {
	return facades.Schema().DropIfExists("sys_user")
}
