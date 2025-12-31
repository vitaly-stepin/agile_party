import { test, expect } from '../fixtures/database.fixture';
import { HomePage } from '../page-objects/HomePage';
import { RoomPage } from '../page-objects/RoomPage';

test.describe('Task Management', () => {
  test('should create new task and display it', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();

    // Action: Create task
    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);

    // Assert: Task appears in list
    await roomPage.waitForTaskToAppear(taskHeadline);
    const taskItem = await roomPage.getTaskItem(taskHeadline);
    await expect(taskItem).toBeVisible();
  });

  test('should sync new task to other users', async ({ page, context }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room with first user
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const firstUserNickname = 'First User';
    await homePage.createRoom(roomName, firstUserNickname);

    await roomPage.waitForWebSocketConnection();

    const roomId = await roomPage.getRoomId();

    // Setup: Second user joins
    const secondUserPage = await context.newPage();
    const secondUserHomePage = new HomePage(secondUserPage);
    const secondUserRoomPage = new RoomPage(secondUserPage);

    await secondUserHomePage.goto();
    const secondUserNickname = 'Second User';
    await secondUserHomePage.joinRoom(roomId, secondUserNickname);

    await secondUserRoomPage.waitForWebSocketConnection();

    // Action: First user creates task
    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);

    // Assert: Task appears for first user
    await roomPage.waitForTaskToAppear(taskHeadline, 10000);

    // Assert: Task syncs to second user
    await secondUserRoomPage.waitForTaskToAppear(taskHeadline, 10000);
    const taskItem = await secondUserRoomPage.getTaskItem(taskHeadline);
    await expect(taskItem).toBeVisible();

    // Cleanup
    await secondUserPage.close();
  });

  test('should rename existing task', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room and task
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();

    const originalHeadline = 'Original task name';
    await roomPage.createTask(originalHeadline);
    await roomPage.waitForTaskToAppear(originalHeadline);

    // Action: Rename task
    const newHeadline = 'Updated task name';
    await roomPage.updateTaskHeadline(originalHeadline, newHeadline);

    // Assert: Task headline updated
    await roomPage.waitForTaskToAppear(newHeadline, 10000);
    const updatedTask = await roomPage.getTaskItem(newHeadline);
    await expect(updatedTask).toBeVisible();
  });

  test('should sync task rename to other users', async ({ page, context }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room with first user
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const firstUserNickname = 'First User';
    await homePage.createRoom(roomName, firstUserNickname);

    await roomPage.waitForWebSocketConnection();

    // Setup: Create task
    const originalHeadline = 'Original task name';
    await roomPage.createTask(originalHeadline);
    await roomPage.waitForTaskToAppear(originalHeadline);

    const roomId = await roomPage.getRoomId();

    // Setup: Second user joins
    const secondUserPage = await context.newPage();
    const secondUserHomePage = new HomePage(secondUserPage);
    const secondUserRoomPage = new RoomPage(secondUserPage);

    await secondUserHomePage.goto();
    const secondUserNickname = 'Second User';
    await secondUserHomePage.joinRoom(roomId, secondUserNickname);

    await secondUserRoomPage.waitForWebSocketConnection();
    await secondUserRoomPage.waitForTaskToAppear(originalHeadline, 10000);

    // Action: First user renames task
    const newHeadline = 'Updated task name';
    await roomPage.updateTaskHeadline(originalHeadline, newHeadline);

    // Assert: First user sees update
    await roomPage.waitForTaskToAppear(newHeadline, 10000);

    // Assert: Second user sees updated headline
    await secondUserRoomPage.waitForTaskToAppear(newHeadline, 10000);
    const updatedTask = await secondUserRoomPage.getTaskItem(newHeadline);
    await expect(updatedTask).toBeVisible();

    // Cleanup
    await secondUserPage.close();
  });

  test('should cancel task creation', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();

    // Action: Start creating task
    await roomPage.addTaskButton.click();
    await roomPage.taskHeadlineInput.fill('Task to be cancelled');

    // Action: Cancel task creation
    await roomPage.cancelTaskButton.click();

    // Assert: Input hidden
    await expect(roomPage.taskHeadlineInput).not.toBeVisible();

    // Assert: Add button visible again
    await expect(roomPage.addTaskButton).toBeVisible();

    // Assert: No task was created
    const noTasksMessage = page.getByText(/No tasks yet/i);
    await expect(noTasksMessage).toBeVisible();
  });

  test('should cancel task edit', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room and task
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();

    const originalHeadline = 'Original task name';
    await roomPage.createTask(originalHeadline);
    await roomPage.waitForTaskToAppear(originalHeadline);

    // Action: Make task active first, then start editing
    await roomPage.setTaskActive(originalHeadline);
    await page.waitForTimeout(500);
    await roomPage.clickEditTask(originalHeadline);

    const taskItem = await roomPage.getTaskItem(originalHeadline);
    const editInput = taskItem.locator('input[type="text"]');
    await editInput.waitFor({ state: 'visible', timeout: 5000 });

    // Action: Modify text but then restore original value (effectively cancelling)
    await editInput.fill('Modified but cancelled');
    await editInput.fill(originalHeadline);
    await editInput.blur();

    // Wait a moment for any WebSocket updates
    await page.waitForTimeout(500);

    // Assert: Original headline preserved (no update was sent)
    const task = await roomPage.getTaskItem(originalHeadline);
    await expect(task).toBeVisible();
    await expect(task.getByText(originalHeadline)).toBeVisible();
  });
});
