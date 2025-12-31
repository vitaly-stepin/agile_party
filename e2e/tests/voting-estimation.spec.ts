import { test, expect } from '../fixtures/database.fixture';
import { HomePage } from '../page-objects/HomePage';
import { RoomPage } from '../page-objects/RoomPage';

test.describe('Voting and Estimation', () => {
  test('should set task as active for voting', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room and task
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();

    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);

    // Action: Set task as active
    await roomPage.setTaskActive(taskHeadline);

    // Assert: Task is highlighted/active
    const taskItem = await roomPage.getTaskItem(taskHeadline);
    const activeIndicator = taskItem.getByText(/Active/i);
    await expect(activeIndicator).toBeVisible();
  });

  test('should submit vote and show voted status', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room, task, and activate it
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();

    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);
    await roomPage.setTaskActive(taskHeadline);

    // Action: Vote
    await roomPage.selectVote('5');

    // Assert: Vote status shows
    await expect(roomPage.voteStatusText).toContainText('You voted 5');
  });

  test('should sync vote status to other users', async ({ page, context }) => {
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

    // Verify both users see each other
    await roomPage.waitForUserToAppear(secondUserNickname, 10000);
    await secondUserRoomPage.waitForUserToAppear(firstUserNickname, 10000);

    // Setup: Create and activate task AFTER both users joined
    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);

    // Second user should see the task (WebSocket sync)
    await secondUserRoomPage.waitForTaskToAppear(taskHeadline, 10000);

    // Now set task as active
    await roomPage.setTaskActive(taskHeadline);

    // Action: First user votes
    await roomPage.selectVote('5');

    // Assert: First user sees their vote
    await expect(roomPage.voteStatusText).toContainText('You voted 5');

    // Assert: Second user sees vote progress (waiting message should appear)
    const waitingMessage = secondUserPage.getByText(/Waiting for all users to vote/i);
    await expect(waitingMessage).toBeVisible({ timeout: 10000 });

    // Cleanup
    await secondUserPage.close();
  });

  test('should allow changing vote before reveal', async ({ page }) => {
    const homePage = new HomePage(page);
    const roomPage = new RoomPage(page);

    // Setup: Create room, task, and vote
    await homePage.goto();
    const roomName = `E2E Test Room ${Date.now()}`;
    const nickname = 'Test User';
    await homePage.createRoom(roomName, nickname);

    await roomPage.waitForWebSocketConnection();

    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);
    await roomPage.setTaskActive(taskHeadline);

    // Action: Vote for 5
    await roomPage.selectVote('5');
    await expect(roomPage.voteStatusText).toContainText('You voted 5');

    // Action: Change vote to 8
    await roomPage.selectVote('8');

    // Assert: Vote status updates
    await expect(roomPage.voteStatusText).toContainText('You voted 8');
  });

  test('should disable reveal button until all users vote', async ({ page, context }) => {
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

    // Verify both users see each other
    await roomPage.waitForUserToAppear(secondUserNickname, 10000);
    await secondUserRoomPage.waitForUserToAppear(firstUserNickname, 10000);

    // Setup: Create and activate task AFTER both users joined
    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);

    // Second user should see the task (WebSocket sync)
    await secondUserRoomPage.waitForTaskToAppear(taskHeadline, 10000);

    // Now set task as active
    await roomPage.setTaskActive(taskHeadline);

    // Action: Only first user votes
    await roomPage.selectVote('5');

    // Assert: Reveal button is disabled
    const isEnabled = await roomPage.isRevealButtonEnabled();
    expect(isEnabled).toBe(false);

    // Assert: Waiting message appears
    const waitingMessage = page.getByText(/Waiting for all users to vote/i);
    await expect(waitingMessage).toBeVisible();

    // Cleanup
    await secondUserPage.close();
  });

  test('should reveal votes when all users have voted', async ({ page, context }) => {
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

    // Verify both users see each other
    await roomPage.waitForUserToAppear(secondUserNickname, 10000);
    await secondUserRoomPage.waitForUserToAppear(firstUserNickname, 10000);

    // Setup: Create and activate task AFTER both users joined
    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);

    // Second user should see the task (WebSocket sync)
    await secondUserRoomPage.waitForTaskToAppear(taskHeadline, 10000);

    // Now set task as active
    await roomPage.setTaskActive(taskHeadline);

    // Action: Both users vote
    await roomPage.selectVote('5');
    await secondUserRoomPage.selectVote('8');

    // Wait a moment for votes to sync
    await page.waitForTimeout(1000);

    // Action: First user reveals
    await roomPage.clickRevealVotes();

    // Assert: Both users see revealed votes
    await roomPage.waitForVotesRevealed(10000);
    await secondUserRoomPage.waitForVotesRevealed(10000);

    await expect(roomPage.averageDisplay).toBeVisible();
    await expect(secondUserRoomPage.averageDisplay).toBeVisible();

    // Cleanup
    await secondUserPage.close();
  });

  test('should display average estimation correctly', async ({ page, context }) => {
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

    // Verify both users see each other
    await roomPage.waitForUserToAppear(secondUserNickname, 10000);
    await secondUserRoomPage.waitForUserToAppear(firstUserNickname, 10000);

    // Setup: Create and activate task AFTER both users joined
    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);

    // Second user should see the task (WebSocket sync)
    await secondUserRoomPage.waitForTaskToAppear(taskHeadline, 10000);

    // Now set task as active
    await roomPage.setTaskActive(taskHeadline);

    // Action: Users vote 5 and 8
    await roomPage.selectVote('5');
    await secondUserRoomPage.selectVote('8');

    // Wait for votes to sync
    await page.waitForTimeout(1000);

    // Action: Reveal votes
    await roomPage.clickRevealVotes();

    // Assert: Average is 6.5 (average of 5 and 8)
    await roomPage.waitForVotesRevealed(10000);
    const average = await roomPage.getAverageEstimate();
    expect(average.trim()).toBe('6.5');

    // Cleanup
    await secondUserPage.close();
  });

  test('should sync revealed votes to all users', async ({ page, context }) => {
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

    // Verify both users see each other
    await roomPage.waitForUserToAppear(secondUserNickname, 10000);
    await secondUserRoomPage.waitForUserToAppear(firstUserNickname, 10000);

    // Setup: Create and activate task AFTER both users joined
    const taskHeadline = 'Implement login feature';
    await roomPage.createTask(taskHeadline);
    await roomPage.waitForTaskToAppear(taskHeadline);

    // Second user should see the task (WebSocket sync)
    await secondUserRoomPage.waitForTaskToAppear(taskHeadline, 10000);

    // Now set task as active
    await roomPage.setTaskActive(taskHeadline);

    // Action: Both users vote
    await roomPage.selectVote('5');
    await secondUserRoomPage.selectVote('8');

    // Wait for votes to sync
    await page.waitForTimeout(1000);

    // Action: First user reveals
    await roomPage.clickRevealVotes();

    // Assert: Second user sees revealed votes automatically (WebSocket sync)
    await secondUserRoomPage.waitForVotesRevealed(10000);
    await expect(secondUserRoomPage.averageDisplay).toBeVisible();

    const secondUserAverage = await secondUserRoomPage.getAverageEstimate();
    expect(secondUserAverage.trim()).toBe('6.5');

    // Cleanup
    await secondUserPage.close();
  });
});
