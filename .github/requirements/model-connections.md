I want to create a model connection template page with mock that use current project style and design.

1. requirements:
	- Add new Model Connections into side bar as this mock up show, 
	- make sure it  has the same hover and focus effect as other side bar items.
	mockup:
	.github/mockups/model-connection/model-connections-navbar.png

2. requirements:
	- After clicking on the link it should open the main model connection page which will be placed under src/app/configuration-ui/model-connections

3. requirements:
	- All component should appear in the position as the mockup.
	mockup:
	.github/mockups/model-connection/model-connection-main-page.png

4. build search bar requirements:
	- Search Bar: design the search bar as mockups shown. 
	- Search bar should filter all the records in the table base on the text input.
	mockups:
	.github/mockups/model-connection/search-bar.png
	.github/mockups/model-connection/search-bar-focus.png

5. build model connection table requirements:
	- create model connection table as mockup shown.
	- Table should have pagination support and show 10 records per page.
	- If there is no record in the table it should show no data found message
	- If there is less than 10 records it should keep the pagination as the same position as with 10 records.
	- The Table: Add row and run link hover effect for all the table rows as shown. 
	- click on the run link will show alert message say test connection with id.
	mockups:
	.github/mockups/model-connection/table-row-hover.png => table design with row hover effect
	.github/mockups/model-connection/run-link-hover.png => run link hover effect

6. build table action bar requirements:
	- Action Popup: action bar shows up when click on three dots(...) of each row. 
	- It should have hover effect for each action as images shown. 
	- Click on edit will open the Model Connection Overlay with data filled from the table record. 
	- Click on delete will delete the record from the table.
	- Make sure state management is updated after delete.
	mockups:
	.github/mockups/model-connection/action-bar.png
	.github/mockups/model-connection/action-bar-focus.png

7. build create model connection button requirements
	- Create Model Connection button, Those images are create model connection button and its hover state.
	mockups:
	.github/mockups/model-connection/create-model-connection-button.png
	.github/mockups/model-connection/create-model-connection-hover.png

8. build create model connection overlay requirements:
	- Create Model Connection Overlay, 
	- this overlay show up when click on the model connection design as image shown.
	- The button in the overlay should have the same affect as model Create Model Connection button
	mockups:
	.github/mockups/model-connection/create-model-connection-overlay.png

9. Cancel button behavior and its popup requirements:
	- Cancel button: when click on the cancel button in the overlay.
	- Shown another popup as mockup shown to confirm the cancel action. 
	- Click on the cancel button in the popup will close the overlay and popup. and return to the main page.
	mockups:
	.github/mockups/model-connection/cancel-popup.png => popup view
	.github/mockups/model-connection/cancel-popup-full.png => Implement the popup location as shown in the full image

10. build model connection service requirements:
	- Build model connection service API endpoints follow RESTFul pattern and hardcode mock data to support the front end CRUD operations.
	- Each type of options in the dropdowns should have a dedicated API to fetch and show up options.
	- Should have a endpoint for testing model connection => it should return success message with 200 status code for now

	Model Connection Type:
		id?: number
		name: string
		createAt?: string
		updatedAt?: string
		providerModelName?: string
		modelType?: string
		requestAdaptor?: string
		description?: string
		endpoint?: string
		apiKeyRedacted?: string
		tag?: string
		customRequestHeader?: string
		customRequestBody?: string
		testRequestBody?: string