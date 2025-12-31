import { test, expect } from '../fixtures/database.fixture';
import { HomePage } from '../page-objects/HomePage';
import { RoomPage } from '../page-objects/RoomPage';

test('should access main page, create room, and join room', async ({ page, context }) => {
  const homePage = new HomePage(page);
  const roomPage = new RoomPage(page);

  // Phase 1: Access main page
  await homePage.goto();
  await expect(page).toHaveTitle(/frontend/i);
  await expect(homePage.createRoomTab).toBeVisible();
  await expect(homePage.joinRoomTab).toBeVisible();

  // Phase 2: Create room
  const roomName = `E2E Test Room ${Date.now()}`;
  const creatorNickname = 'Test Creator';
  await homePage.createRoom(roomName, creatorNickname);

  // Verify navigation to room
  await expect(page).toHaveURL(/\/room\/[a-f0-9-]+/);

  // Phase 3: Verify room loaded
  await expect(roomPage.roomTitle).toContainText(roomName);
  await roomPage.waitForWebSocketConnection();
  await roomPage.waitForUserToAppear(creatorNickname);

  // Phase 4: Join as second user
  const roomId = await roomPage.getRoomId();

  // Open new browser context for second user
  const secondUserPage = await context.newPage();
  const secondUserHomePage = new HomePage(secondUserPage);
  const secondUserRoomPage = new RoomPage(secondUserPage);

  await secondUserHomePage.goto();

  const joinerNickname = 'Test Joiner';
  await secondUserHomePage.joinRoom(roomId, joinerNickname);

  // Verify navigation to room
  await expect(secondUserPage).toHaveURL(/\/room\/[a-f0-9-]+/);

  // Phase 5: Verify WebSocket real-time sync
  // Second user's view: verify creator is visible
  await expect(secondUserRoomPage.roomTitle).toContainText(roomName);
  await secondUserRoomPage.waitForWebSocketConnection();
  await secondUserRoomPage.waitForUserToAppear(creatorNickname, 10000);

  // First user's view: wait for joiner to appear
  await roomPage.waitForUserToAppear(joinerNickname, 10000);

  // Verify both users see the same room name
  await expect(roomPage.roomTitle).toContainText(roomName);
  await expect(secondUserRoomPage.roomTitle).toContainText(roomName);

  // Cleanup: close second user context
  await secondUserPage.close();
});
