WITH r AS(
	SELECT 
		movieId, 
		AVG(rating) AS avgRating, 
		COUNT(rating) AS nRatings 
	FROM movierating 
	GROUP BY movieId
)
UPDATE movies m
JOIN r ON m.movieId = r.movieId
SET m.rating=r.avgRating, m.ratingCount=r.nRatings
	