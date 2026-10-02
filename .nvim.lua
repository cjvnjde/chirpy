vim.g.db = "postgres://postgres:postgres@localhost:5432/chirpy?sslmode=disable"

vim.api.nvim_create_autocmd("VimEnter", {
	once = true,
	callback = function()
		local editor_tab = vim.api.nvim_get_current_tabpage()

		-- Tab 2: terminal
		vim.cmd("tablast")
		vim.cmd("tabnew")
		vim.cmd("terminal")

		-- Tab 3: database UI
		vim.cmd("tabnew")
		vim.cmd("DBUI")

		-- Start in the original editor tab
		vim.api.nvim_set_current_tabpage(editor_tab)
	end,
})
