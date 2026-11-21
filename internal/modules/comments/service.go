package comments

import (
	"context"
	"errors"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"gopress/internal/model"
)

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrParentNotFound  = errors.New("parent comment not found")
	ErrCommentNotFound = errors.New("comment not found")
	qqPattern          = regexp.MustCompile(`^[1-9][0-9]{4,11}$`)
	qqEmailDomains     = map[string]struct{}{
		"qq.com": {}, "vip.qq.com": {}, "foxmail.com": {},
	}
)

// ValidationError carries a message that is safe to show to the commenter.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CreateInput struct {
	Nickname string
	Email    string
	QQ       string
	Content  string
	ParentID *uint
}

type ListParams struct {
	Page     int
	PageSize int
}

type ListResult struct {
	Items        []model.CommentDTO
	Total        int64
	CommentCount int64
	Page         int
	PageSize     int
	TotalPages   int
}

type AdminListParams struct {
	Page     int
	PageSize int
	Keyword  string
	PostID   *uint
}

func (s *Service) publishedPost(ctx context.Context, slug string) (model.Post, error) {
	var post model.Post
	err := s.db.WithContext(ctx).
		Where("slug = ? AND status = ? AND published_at IS NOT NULL AND published_at <= ?",
			slug, model.StatusPublished, time.Now()).
		First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Post{}, ErrPostNotFound
	}
	return post, err
}

func (s *Service) List(ctx context.Context, slug string, params ListParams) (ListResult, error) {
	post, err := s.publishedPost(ctx, slug)
	if err != nil {
		return ListResult{}, err
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	var rootTotal int64
	if err := s.db.WithContext(ctx).Model(&model.Comment{}).
		Where("post_id = ? AND parent_id IS NULL", post.ID).Count(&rootTotal).Error; err != nil {
		return ListResult{}, err
	}
	var commentCount int64
	if err := s.db.WithContext(ctx).Model(&model.Comment{}).
		Where("post_id = ?", post.ID).Count(&commentCount).Error; err != nil {
		return ListResult{}, err
	}
	roots := make([]model.Comment, 0, params.PageSize)
	if err := s.db.WithContext(ctx).Where("post_id = ? AND parent_id IS NULL", post.ID).
		Order("created_at DESC").Order("id DESC").
		Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Find(&roots).Error; err != nil {
		return ListResult{}, err
	}
	replies := make([]model.Comment, 0)
	if len(roots) > 0 {
		rootIDs := make([]uint, 0, len(roots))
		for _, root := range roots {
			rootIDs = append(rootIDs, root.ID)
		}
		if err := s.db.WithContext(ctx).Where("post_id = ? AND root_id IN ?", post.ID, rootIDs).
			Order("created_at ASC").Order("id ASC").Find(&replies).Error; err != nil {
			return ListResult{}, err
		}
	}
	totalPages := 0
	if rootTotal > 0 {
		totalPages = int((rootTotal + int64(params.PageSize) - 1) / int64(params.PageSize))
	}
	return ListResult{
		Items: BuildTree(roots, replies), Total: rootTotal, CommentCount: commentCount,
		Page: params.Page, PageSize: params.PageSize, TotalPages: totalPages,
	}, nil
}

func (s *Service) Create(ctx context.Context, slug string, input CreateInput) (model.CommentDTO, error) {
	post, err := s.publishedPost(ctx, slug)
	if err != nil {
		return model.CommentDTO{}, err
	}
	input, err = normalizeInput(input)
	if err != nil {
		return model.CommentDTO{}, err
	}

	comment := model.Comment{
		PostID: post.ID, Nickname: input.Nickname, Email: input.Email,
		QQ: input.QQ, Content: input.Content,
	}
	if input.ParentID != nil {
		if *input.ParentID == 0 {
			return model.CommentDTO{}, &ValidationError{Message: "回复的评论不存在"}
		}
		var parent model.Comment
		if err := s.db.WithContext(ctx).Where("id = ? AND post_id = ?", *input.ParentID, post.ID).
			First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return model.CommentDTO{}, ErrParentNotFound
			}
			return model.CommentDTO{}, err
		}
		comment.ParentID = &parent.ID
		if parent.RootID != nil {
			comment.RootID = parent.RootID
		} else {
			comment.RootID = &parent.ID
		}
	}
	if err := s.db.WithContext(ctx).Create(&comment).Error; err != nil {
		return model.CommentDTO{}, err
	}
	return ToPublicDTO(comment), nil
}

func (s *Service) AdminList(ctx context.Context, params AdminListParams) ([]model.AdminCommentDTO, int64, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	apply := func(query *gorm.DB) *gorm.DB {
		if keyword := strings.TrimSpace(params.Keyword); keyword != "" {
			like := "%" + keyword + "%"
			query = query.Where(
				"comments.nickname LIKE ? OR comments.content LIKE ? OR comments.email LIKE ? OR comments.qq LIKE ?",
				like, like, like, like,
			)
		}
		if params.PostID != nil {
			query = query.Where("comments.post_id = ?", *params.PostID)
		}
		return query
	}
	var total int64
	if err := apply(s.db.WithContext(ctx).Table("comments").
		Joins("JOIN posts ON posts.id = comments.post_id")).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	type adminRow struct {
		ID             uint
		PostID         uint
		PostTitle      string
		PostSlug       string
		ParentID       *uint
		ParentNickname string
		Nickname       string
		Email          string
		QQ             string
		Content        string
		CreatedAt      time.Time
	}
	rows := make([]adminRow, 0, params.PageSize)
	err := apply(s.db.WithContext(ctx).Table("comments").
		Joins("JOIN posts ON posts.id = comments.post_id").
		Joins("LEFT JOIN comments AS parent ON parent.id = comments.parent_id")).
		Select(`comments.id, comments.post_id, posts.title AS post_title, posts.slug AS post_slug,
			comments.parent_id, COALESCE(parent.nickname, '') AS parent_nickname,
			comments.nickname, comments.email, comments.qq, comments.content, comments.created_at`).
		Order("comments.created_at DESC").Order("comments.id DESC").
		Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	items := make([]model.AdminCommentDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.AdminCommentDTO{
			ID: row.ID, PostID: row.PostID, PostTitle: row.PostTitle, PostSlug: row.PostSlug,
			ParentID: row.ParentID, ParentNickname: row.ParentNickname,
			Nickname: row.Nickname, Email: row.Email, QQ: row.QQ,
			Content: row.Content, CreatedAt: row.CreatedAt,
		})
	}
	return items, total, nil
}

// Delete removes a comment together with every descendant reply so the public
// thread never keeps orphaned branches.
func (s *Service) Delete(ctx context.Context, id uint) error {
	var existing model.Comment
	if err := s.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}
	ids := []uint{id}
	frontier := []uint{id}
	for depth := 0; depth < 32 && len(frontier) > 0; depth++ {
		var children []uint
		if err := s.db.WithContext(ctx).Model(&model.Comment{}).
			Where("parent_id IN ?", frontier).Pluck("id", &children).Error; err != nil {
			return err
		}
		ids = append(ids, children...)
		frontier = children
	}
	return s.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Comment{}).Error
}

func normalizeInput(input CreateInput) (CreateInput, error) {
	input.Nickname = strings.TrimSpace(input.Nickname)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.QQ = strings.TrimSpace(input.QQ)
	input.Content = strings.TrimSpace(input.Content)

	if n := utf8.RuneCountInString(input.Nickname); n < 1 || n > 40 {
		return CreateInput{}, &ValidationError{Message: "昵称需为 1-40 个字符"}
	}
	if n := utf8.RuneCountInString(input.Content); n < 1 || n > 2000 {
		return CreateInput{}, &ValidationError{Message: "评论内容需为 1-2000 个字符"}
	}
	parsed, err := mail.ParseAddress(input.Email)
	if err != nil || parsed.Address != input.Email {
		return CreateInput{}, &ValidationError{Message: "请输入有效的 QQ 邮箱"}
	}
	at := strings.LastIndex(input.Email, "@")
	if at <= 0 {
		return CreateInput{}, &ValidationError{Message: "请输入有效的 QQ 邮箱"}
	}
	domain := input.Email[at+1:]
	if _, ok := qqEmailDomains[domain]; !ok {
		return CreateInput{}, &ValidationError{Message: "请使用 QQ 邮箱（@qq.com、@vip.qq.com 或 @foxmail.com）"}
	}
	if input.QQ == "" && qqPattern.MatchString(input.Email[:at]) {
		input.QQ = input.Email[:at]
	}
	if input.QQ != "" && !qqPattern.MatchString(input.QQ) {
		return CreateInput{}, &ValidationError{Message: "QQ 号需为 5-12 位数字"}
	}
	return input, nil
}

func AvatarURL(qq string) string {
	if qq == "" {
		return ""
	}
	return "https://q1.qlogo.cn/g?b=qq&nk=" + url.QueryEscape(qq) + "&s=100"
}

func ToPublicDTO(comment model.Comment) model.CommentDTO {
	return model.CommentDTO{
		ID: comment.ID, PostID: comment.PostID, ParentID: comment.ParentID,
		Nickname: comment.Nickname, Avatar: AvatarURL(comment.QQ),
		Content: comment.Content, CreatedAt: comment.CreatedAt,
		Replies: make([]model.CommentDTO, 0),
	}
}

// BuildTree turns flat roots and descendants into an ordered tree. Replies that
// reference a missing parent are dropped rather than promoted to roots.
func BuildTree(roots, replies []model.Comment) []model.CommentDTO {
	type node struct {
		dto      model.CommentDTO
		children []*node
	}
	nodes := make(map[uint]*node, len(roots)+len(replies))
	ordered := make([]*node, 0, len(roots))
	for _, root := range roots {
		item := &node{dto: ToPublicDTO(root)}
		nodes[root.ID] = item
		ordered = append(ordered, item)
	}
	for _, reply := range replies {
		nodes[reply.ID] = &node{dto: ToPublicDTO(reply)}
	}
	for _, reply := range replies {
		if reply.ParentID == nil {
			continue
		}
		if parent, ok := nodes[*reply.ParentID]; ok {
			parent.children = append(parent.children, nodes[reply.ID])
		}
	}
	var build func(*node) model.CommentDTO
	build = func(item *node) model.CommentDTO {
		dto := item.dto
		dto.Replies = make([]model.CommentDTO, 0, len(item.children))
		for _, child := range item.children {
			dto.Replies = append(dto.Replies, build(child))
		}
		return dto
	}
	result := make([]model.CommentDTO, 0, len(ordered))
	for _, item := range ordered {
		result = append(result, build(item))
	}
	return result
}
