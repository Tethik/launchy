package main

// Packaging this with the compiled binary due to lazyiness. Much easier to install that way.
// TODO: (Potential) move this to an external file
const stylesheet = `
.window {
	background-color: black;
	min-height: 75px;
  }
  
  .search_bar {
	/* min-height: 40px; */
	margin: 5px;
  }
  
  .result_row {
	min-height: 50px;
  }
  
  .app_icon {
	min-height: 64px;
  }
  
  .result_list {
	padding: 5px;
  }
`
