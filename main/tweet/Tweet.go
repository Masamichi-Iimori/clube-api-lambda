package tweet

// User つぶやいたユーザ情報
type User struct {
	ID         int64  `dynamo:"id" json:"id"`
	Name       string `dynamo:"name" json:"name"`
	ScreenName string `dynamo:"screen_name" json:"screen_name"`
}

// Tweet 参加を募集するツイート
type Tweet struct {
	ID        int64    `dynamo:"tweet_id" json:"tweet_id"`
	FullText  string   `dynamo:"full_text" json:"full_text"`
	TweetedAt int64    `dynamo:"tweeted_at" json:"tweeted_at"`
	User      User     `dynamo:"user" json:"user"`
	Position  []string `dynamo:"position" json:"positions"`
	MediaURLs []string `dynamo:"media_url" json:"media_url"`
	IsClub    bool     `dynamo:"is_club" json:"is_club"`
}

// Tweets 構造体のスライス
type Tweets []Tweet

// 以下インタフェースを渡してTweetedAtでソート可能にする
func (t Tweets) Len() int {
	return len(t)
}

func (t Tweets) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// 新しい順にソート
func (t Tweets) Less(i, j int) bool {
	return t[i].TweetedAt > t[j].TweetedAt
}
