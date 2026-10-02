local project_dir = vim.fn.fnamemodify(debug.getinfo(1, "S").source:sub(2), ":h")
vim.g.db_ui_save_location = project_dir .. "/.dbui"

vim.g.db = "postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"

vim.api.nvim_create_autocmd("VimEnter", {
	once = true,
	callback = function()
		local editor_tab = vim.api.nvim_get_current_tabpage()

		vim.cmd("tablast")
		vim.cmd("tabnew")
		vim.cmd("DBUI")

		vim.api.nvim_set_current_tabpage(editor_tab)
	end,
})
